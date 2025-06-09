import {ApolloServer} from '@apollo/server';
import {expressMiddleware} from '@apollo/server/express4';
import express from 'express';
import {ApolloGateway, IntrospectAndCompose, RemoteGraphQLDataSource} from '@apollo/gateway';
import fs from 'fs/promises';
import assert from 'assert';

const startingTime = new Date()

const useEnv = (key) => {
  const v = process.env[key];
  assert(v, `env ${key} must be specified`)
  return v
};

// const cfgMap = yaml.load(await fs.readFile(useEnv("SUPERGRAPH_CONFIG"), 'utf8'));

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

  didReceiveResponse({ response, context }) {
    const x = response.http.headers.get('set-cookie');
    if (!x) return response;
    // console.log("context", context)
    context.res.setHeader('set-cookie', x);
    context.id = response.id;
    return response;
  }
}


const superGraphSchema = await fs.readFile(useEnv("SUPERGRAPH_SCHEMA_FILE"));

const gateway = new ApolloGateway({
  supergraphSdl: superGraphSchema.toString(),
  // supergraphSdl: new IntrospectAndCompose({
  //   subgraphs: cfgMap.serviceList,
  // }),
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

app.get("/healthz", (req, res) => {
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
    process.exit(1)
  }

  const timeTaken = ((new Date().getTime() - startingTime)/1e3).toFixed(2)

  console.log(String.raw`
          **                                                               
        ****                                                               
       ****      *                                                          
      ****     *****       ▗▖ ▗▖▗▖    ▗▄▖ ▗▖ ▗▖▗▄▄▄ ▗▖   ▗▄▄▄▖▗▄▄▄▖▗▄▄▄▖    
    ****     *********     ▐▌▗▞▘▐▌   ▐▌ ▐▌▐▌ ▐▌▐▌  █▐▌     █    █  ▐▌       
   ****     ************   ▐▛▚▖ ▐▌   ▐▌ ▐▌▐▌ ▐▌▐▌  █▐▌     █    █  ▐▛▀▀▘    
    ****     *********     ▐▌ ▐▌▐▙▄▄▖▝▚▄▞▘▝▚▄▞▘▐▙▄▄▀▐▙▄▄▖▗▄█▄▖  █  ▐▙▄▄▖    
      ****     *****                                                        
       ****      *            🚀 running in ${timeTaken}s                   
        ****                                                               
          **                                                               
                                                                            
  `)
  console.log(`🚀 Federation Gateway ready at :${port}`);
})

