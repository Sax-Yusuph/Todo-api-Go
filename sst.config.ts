/// <reference path="./.sst/platform/config.d.ts" />

export default $config({
  app(input) {
    return {
      name: "todo",
      removal: input?.stage === "production" ? "retain" : "remove",
      protect: ["production"].includes(input?.stage),
      home: "aws",
      providers: {
        aws: {
          profile: ["production"].includes(input?.stage)
            ? "basecliff-prod"
            : "basecliff-dev",
          region: "eu-west-2",
        },
      },
    };
  },
  async run() {
    new sst.aws.Function("GoFunction", {
      url: true,
      runtime: "go",
      handler: "./src",
    });
  },
});
