/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
    // By default, Docusaurus generates a sidebar from the docs folder structure
    tutorialSidebar: [
      { type: "autogenerated", dirName: "." },
  
      {
        type: "category",
        label: "Resources",
        collapsed: false,
        items: [
          {
            type: "link",
            label: "Tutorials",
            href: "https://tutorials.cosmos.network",
          },
          {
            type: "link",
            label: "SDK API Reference",
            href: "https://pkg.go.dev/github.com/cosmos/cosmos-sdk",
          },
          {
            type: "link",
            label: "REST API Spec",
            href: "https://docs.cosmos.network/swagger/",
          },
          {
            type: "link",
            label: "Awesome Cosmos",
            href: "https://github.com/cosmos/awesome-cosmos",
          },
          {
            type: "link",
            label: "Support",
            href: "https://github.com/orgs/cosmos/discussions",
          },
        ],
      },
    ],
  };
  
  module.exports = sidebars;