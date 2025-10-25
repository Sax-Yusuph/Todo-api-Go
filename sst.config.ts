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
        cloudflare: true,
      },
    };
  },
  async run() {
    const jwt = new sst.Secret("JWTSecret");
    const redis = new sst.Secret("RedisUrl");

    const domain = (() => {
      const domains: Record<string, string> = {
        production: "todo.basecliff.com",
        dev: "todo-dev.basecliff.com",
      };

      return domains[$app.stage] || `todo-${$app.stage}.basecliff.com`;
    })();

    const router = new sst.aws.Router("TodoRouter", {
      domain: {
        name: domain,
        dns: sst.cloudflare.dns(),
      },
    });

    new sst.aws.Function("TodoFunction", {
      url: {
        router: {
          instance: router,
          domain,
        },
      },
      runtime: "go",
      handler: "./src",
      link: [jwt, redis],
      streaming: true,
      memory: "128 MB",
    });
  },
});
