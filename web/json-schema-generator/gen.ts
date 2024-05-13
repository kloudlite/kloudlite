import tjs from 'typescript-json-schema';
import fs from 'fs';

const types: string[] = ['ConsoleUpdateAppMutationVariables'];

async function fake(files: string[], types: string[] = []) {
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

  const datas: {
    [key: string]: tjs.Definition;
  } = {};

  userSymbols.forEach((sym) => {
    if (types.includes(sym)) {
      console.log(sym);
      const jsonSchema = generator.getSchemaForSymbol(sym);
      datas[sym] = jsonSchema;
    }
  });

  return datas;
}

// @ts-ignore
// just to make it work in node
global.location = new URL('https://kloudlite.io');

(async () => {
  const data = await fake(['src/generated/gql/server.ts'], types);

  console.log(data);

  fs.writeFileSync(
    './json-schema-generator/schema.js',
    `// Generated file, Generated for validators

const schema = ${JSON.stringify(data, null, 2)};

export default schema;`
  );
})();
