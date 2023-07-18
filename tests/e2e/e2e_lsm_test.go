package e2e

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
)

func (s *IntegrationTestSuite) testLSM() {
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

	validatorA := s.chainA.validators[0]
	validatorAAddr := validatorA.keyInfo.GetAddress()

	validatorAddressA := sdk.ValAddress(validatorAAddr).String()

	// Set parameters (global liquid staking cap, validator liquid staking cap, validator bond factor)
	s.writeLiquidStakingParamsUpdateProposal(s.chainA)
	submitGovFlags := []string{"param-change", configFile(proposalLSMParamUpdateFilename)}
	depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes"}

	// gov proposing LSM parameters (global liquid staking cap, validator liquid staking cap, validator bond factor)
	s.T().Logf("Proposal number: %d", proposalCounter)
	s.T().Logf("Submitting, deposit and vote legacy Gov Proposal: Set parameters (global liquid staking cap, validator liquid staking cap, validator bond factor)")
	s.runGovProcess(chainEndpoint, validatorAAddr.String(), proposalCounter, paramtypes.ProposalTypeChange, submitGovFlags, depositGovFlags, voteGovFlags, "vote", false)

	// query the proposal status and new fee
	s.Require().Eventually(
		func() bool {
			proposal, err := queryGovProposal(chainEndpoint, proposalCounter)
			s.Require().NoError(err)
			return proposal.GetProposal().Status == gov.StatusPassed
		},
		15*time.Second,
		5*time.Second,
	)

	s.Require().Eventually(
		func() bool {
			stakingParams, err := queryStakingParams(chainEndpoint)
			s.T().Logf("After LSM parameters update proposal")
			s.Require().NoError(err)

			s.Require().Equal(stakingParams.Params.GlobalLiquidStakingCap, sdk.NewDecWithPrec(30, 2))
			s.Require().Equal(stakingParams.Params.ValidatorLiquidStakingCap, sdk.NewDecWithPrec(100, 2))
			s.Require().Equal(stakingParams.Params.ValidatorBondFactor, sdk.NewDec(-1))

			return true
		},
		15*time.Second,
		5*time.Second,
	)
	delegatorAddress := s.chainA.genesisAccounts[2].keyInfo.GetAddress().String()

	fees := sdk.NewCoin(uatomDenom, sdk.NewInt(1))

	// Validator bond
	s.executeValidatorBond(s.chainA, 0, validatorAddressA, validatorAAddr.String(), gaiaHomePath, fees.String())

	// Validate validator bond successful
	selfBondedShares := sdk.ZeroDec()
	s.Require().Eventually(
		func() bool {
			res, err := queryDelegation(chainEndpoint, validatorAddressA, validatorAAddr.String())
			delegation := res.GetDelegationResponse().GetDelegation()
			selfBondedShares = delegation.Shares
			isValidatorBond := delegation.ValidatorBond
			s.Require().NoError(err)

			return isValidatorBond == true
		},
		20*time.Second,
		5*time.Second,
	)

	delegationAmount := sdk.NewInt(500000000)
	delegation := sdk.NewCoin(uatomDenom, delegationAmount) // 500 atom

	// Alice delegate uatom to Validator A
	s.executeDelegate(s.chainA, 0, delegation.String(), validatorAddressA, delegatorAddress, gaiaHomePath, fees.String())

	// Validate delegation successful
	s.Require().Eventually(
		func() bool {
			res, err := queryDelegation(chainEndpoint, validatorAddressA, delegatorAddress)
			amt := res.GetDelegationResponse().GetDelegation().GetShares()
			s.Require().NoError(err)

			return amt.Equal(sdk.NewDecFromInt(delegationAmount))
		},
		20*time.Second,
		5*time.Second,
	)

	// Tokenize shares
	tokenizeAmount := sdk.NewInt(200000000)
	tokenize := sdk.NewCoin(uatomDenom, tokenizeAmount) // 200 atom
	s.executeTokenizeShares(s.chainA, 0, tokenize.String(), validatorAddressA, delegatorAddress, gaiaHomePath, fees.String())

	// Validate delegation reduced
	s.Require().Eventually(
		func() bool {
			res, err := queryDelegation(chainEndpoint, validatorAddressA, delegatorAddress)
			amt := res.GetDelegationResponse().GetDelegation().GetShares()
			s.Require().NoError(err)

			return amt.Equal(sdk.NewDecFromInt(delegationAmount.Sub(tokenizeAmount)))
		},
		20*time.Second,
		5*time.Second,
	)

	// Validate balance increased
	recordId := int(1)
	shareDenom := fmt.Sprintf("%s/%s", strings.ToLower(validatorAddressA), strconv.Itoa(recordId))
	s.Require().Eventually(
		func() bool {
			res, err := getSpecificBalance(chainEndpoint, delegatorAddress, shareDenom)
			s.Require().NoError(err)
			return res.Amount.Equal(tokenizeAmount)
		},
		20*time.Second,
		5*time.Second,
	)

	// Bank send LSM token
	sendAmount := sdk.NewCoin(shareDenom, tokenizeAmount)
	s.execBankSend(s.chainA, 0, delegatorAddress, validatorAAddr.String(), sendAmount.String(), standardFees.String(), false)

	// Validate tokens are sent properly
	s.Require().Eventually(
		func() bool {
			afterSenderShareDenomBalance, err := getSpecificBalance(chainEndpoint, delegatorAddress, shareDenom)
			s.Require().NoError(err)

			afterRecipientShareDenomBalance, err := getSpecificBalance(chainEndpoint, validatorAAddr.String(), shareDenom)
			s.Require().NoError(err)

			decremented := afterSenderShareDenomBalance.IsZero()
			incremented := afterRecipientShareDenomBalance.IsEqual(sendAmount)

			return decremented && incremented
		},
		time.Minute,
		5*time.Second,
	)

	// transfer reward ownership
	s.executeTransferTokenizeShareRecord(s.chainA, 0, strconv.Itoa(recordId), delegatorAddress, validatorAAddr.String(), gaiaHomePath, standardFees.String())

	// Validate ownership transferred correctly
	s.Require().Eventually(
		func() bool {
			record, err := queryTokenizeShareRecordById(chainEndpoint, recordId)
			s.Require().NoError(err)
			return record.Owner == validatorAAddr.String()
		},
		time.Minute,
		5*time.Second,
	)

	// TODO: IBC transfer LSM token

	// Redeem tokens for shares
	s.executeRedeemShares(s.chainA, 0, sendAmount.String(), validatorAAddr.String(), gaiaHomePath, fees.String())

	// check redeem success
	s.Require().Eventually(
		func() bool {
			delegationRes, err := queryDelegation(chainEndpoint, validatorAddressA, validatorAAddr.String())
			delegation := delegationRes.GetDelegationResponse().GetDelegation()
			s.Require().NoError(err)

			balanceRes, err := getSpecificBalance(chainEndpoint, delegatorAddress, shareDenom)
			s.Require().NoError(err)
			return balanceRes.Amount.IsZero() && delegation.Shares.GT(selfBondedShares)
		},
		20*time.Second,
		5*time.Second,
	)
}
