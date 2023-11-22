import fs from 'fs';
import jsf from 'json-schema-faker';
import tjs from 'typescript-json-schema';
// @ts-ignore
import a from '@faker-js/faker';

const types: string[] = [
  'ConsoleListProjectsQuery',
  'ConsoleListClustersQuery',
  'ConsoleListRepoQuery',
  'ConsoleListCredQuery',
  'ConsoleListProviderSecretsQuery',
  'ConsoleListNodePoolsQuery',
  'ConsoleListVpnDevicesQuery',
  'ConsoleListDomainsQuery',
  'ConsoleListWorkspacesQuery',
  'ConsoleListEnvironmentsQuery',
  'ConsoleListAppsQuery',
  'ConsoleListConfigsQuery',
  'ConsoleListSecretsQuery',
  'ConsoleListManagedServicesQuery',
  'ConsoleListManagedResourceQuery',
];

async function fake(files: string[], types: string[] = []) {
  jsf.extend('faker', () => a);
  jsf.option({
    maxItems: 4,
    minItems: 4,
    alwaysFakeOptionals: true,
    ignoreProperties: [
      'output',
      'resources',
      // 'syncStatus',
      'annotations',
      'recordVersion',
      'lastReconcileTime',
      'tolerations',
    ],
  });

  const program = tjs.getProgramFromFiles(files, { lib: ['esnext'] });
  const validationKeywords = ['faker'];
  const generator = tjs.buildGenerator(
    program,
    { validationKeywords, required: true, include: files },
    files
  );
  if (!generator) {
    throw new Error('Failed to create generator');
  }
  const userSymbols = generator.getUserSymbols();
  const datas = {};
  const fakeDatas = await Promise.all(
    userSymbols.map((sym) => {
      if (types && !types.includes(sym)) {
        return null;
      }
      const jsonSchema = generator.getSchemaForSymbol(sym);
      return jsf.resolve(jsonSchema, []);
    })
  );
  for (let i = 0; i < fakeDatas.length; i += 1) {
    if (fakeDatas[i]) {
      // @ts-ignore
      datas[userSymbols[i]] = fakeDatas[i];
    }
  }
  return datas;
}

// @ts-ignore
// just to make it work in node
global.location = new URL(
  'file:///home/runner/work/nextjs-graphql/nextjs-graphql/src/generated/gql/server.ts'
);

(async () => {
  const data = await fake(['src/generated/gql/server.ts'], types);

  fs.writeFileSync(
    './fake-data-generator/fake.js',
    `const fake = ${JSON.stringify(data, null, 2)};
     export default fake;
`
  );
})();
