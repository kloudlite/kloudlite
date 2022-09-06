import { ApolloServer } from 'apollo-server';
import { ApolloGateway, RemoteGraphQLDataSource, IntrospectAndCompose } from '@apollo/gateway';
import fs from 'fs/promises';
import yaml from 'js-yaml';
import assert from 'assert';
import path from 'path';

const useEnv = (key) => {
  const v = process.env[key];
  assert(v, `env ${key} must be specified`)
  return v
};

const cfgMap = yaml.load(await fs.readFile(useEnv("SUPERGRAPH_CONFIG"), 'utf8'));

// const supergraphSdl = (
//   await fs.readFile(useEnv('SUPERGRAPH_CONFIG'), 'utf8')
// ).toString();

class CustomDataSource extends RemoteGraphQLDataSource {
  // eslint-disable-next-line class-methods-use-this
  willSendRequest({ request, context }) {
    // console.log("--------------")
    // console.log("ctx.headers: ", context?.req?.headers)
    // console.log("--------------")
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
  context: async ({ req, res }) => {
    return { req, res };
  },
});

const port = useEnv("PORT")
const { url } = await server.listen({ port });
console.log(`🚀 Federation Gateway ready at ${url}`);
