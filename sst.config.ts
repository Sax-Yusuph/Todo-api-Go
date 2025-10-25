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
    const jwt = new sst.Secret("JWTSecret");
    const redis = new sst.Secret("RedisUrl");
    new sst.aws.Function("TodoFunction", {
      url: true,
      runtime: "go",
      handler: "./src",
      link: [jwt, redis],
      streaming: true,
    });
  },
});
