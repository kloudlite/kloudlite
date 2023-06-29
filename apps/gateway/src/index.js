import { ApolloServer } from '@apollo/server';
import { startStandaloneServer } from '@apollo/server/standalone';
import { expressMiddleware } from '@apollo/server/express4';
import express from 'express';
import { ApolloGateway, RemoteGraphQLDataSource, IntrospectAndCompose } from '@apollo/gateway';
import fs from 'fs/promises';
import yaml from 'js-yaml';
import assert from 'assert';

const useEnv = (key) => {
  const v = process.env[key];
  assert(v, `env ${key} must be specified`)
  return v
};

// const cfgMap = yaml.load(await fs.readFile(useEnv("SUPERGRAPH_CONFIG"), 'utf8'));
const cfgMap = {
  serviceList: [
    { name: 'auth-api', url: 'http://auth-api.kl-core.svc.cluster.local/query' },
    { name: 'infra-api', url: 'http://infra-api.kl-core.svc.cluster.local/query' },
    { name: 'console-api', url: 'http://console-api.kl-core.svc.cluster.local/query' },
    { name: 'finance-api', url: 'http://finance-api.kl-core.svc.cluster.local/query' },
    { name: 'message-office-api', url: 'http://message-office-api.kl-core.svc.cluster.local/query' },
  ]
}

// const supergraphSdl = (
//   await fs.readFile(useEnv('SUPERGRAPH_CONFIG'), 'utf8')
// ).toString();
//
class CustomDataSource extends RemoteGraphQLDataSource {
  // eslint-disable-next-line class-methods-use-this
  willSendRequest({ request, context }) {
    // console.log("context.req.headers", context?.req?.headers)
    if (context && context.req && context.req.headers) {
      Object.entries(context.req.headers).forEach(([key, value]) => {
        request.http.headers.set(key, value);
      });
    }
    return request;
  }

  // eslint-disable-next-line class-methods-use-this
  didReceiveResponse({ response, context }) {
    const x = response.http.headers.get('set-cookie');
    if (!x) return response;
    // console.log("context", context)
    context.res.setHeader('set-cookie', x);
    context.id = response.id;
    return response;
  }
}

const gateway = new ApolloGateway({
  supergraphSdl: new IntrospectAndCompose({
    subgraphs: cfgMap.serviceList,
  }),
  buildService({ name, url }) {
    return new CustomDataSource({ name, url });
  },
});

const server = new ApolloServer({
  cors: {
    origin: new RegExp(
      '(https://studio.apollographql.com)|(https?://localhost:[43]000)'
    ),
    credentials: true,
  },
  gateway,
  // plugins: [graphqlExecutionLogger],
  // context: async ({ req, res }) => {
  //   return { req, res };
  // },
});

const app = express()
await server.start()

app.get("/health", (req, res) => {
  return res.sendStatus(200);
})

app.use('/', express.json(), expressMiddleware(server, {
  context: async ({ req, res }) => {
    return { req, res };
  },
}));

const port = useEnv("PORT")
app.listen(port, (err) => {
  if (err) {
    console.error("failed to start express server")
  }
  console.log(`🚀 Federation Gateway ready at :${port}`);
})
