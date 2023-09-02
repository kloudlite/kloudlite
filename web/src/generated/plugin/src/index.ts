import { PluginFunction, Types } from '@graphql-codegen/plugin-helpers';
import { GraphQLSchema, isEnumType } from 'graphql';
import './dump';

export const plugin: PluginFunction = (
  schema: GraphQLSchema,
  documents: Types.DocumentFile[],
  config: any
) => {
  console.log('hllow');
  const allTypes = Object.values(schema.getTypeMap());
  const enumTypes = allTypes.filter(isEnumType);

  console.log(enumTypes, 'helllo');

  // const result = enumTypes
  //   .map((enumType) => {
  //     const values = enumType
  //       .getValues()
  //       .map((value) => `'${value.name}'`)
  //       .join(' | ');
  //     return `export type ${enumType.name} = ${values} | string;`;
  //   })
  //   .join('\n');

  return { prepend: [], content: '' };
};
