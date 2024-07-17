/* eslint-disable no-continue */
import { buildASTSchema, parse, GraphQLNamedType } from 'graphql';
import { useMemo } from 'react';
// @ts-ignore
import a from './sdl.graphql';

// const typeDefs = fs.readFileSync('./gql/sdl.graphql');
const typeDefs = a;

const astDocument = parse(typeDefs.toString());
const schema = buildASTSchema(astDocument);

function validateInput(
  input: any,
  graphqlInputType: GraphQLNamedType | undefined
) {
  let validationErrors: string[] = [];

  // @ts-ignore
  const fields = graphqlInputType?._fields || [];

  // eslint-disable-next-line guard-for-in, no-restricted-syntax
  for (const field in fields) {
    let fieldType = fields[field].type;
    // const fieldAstNode = fields[field].astNode;

    if (!input) {
      validationErrors.push(`${field} is ${fieldType.constructor.name}`);
      continue;
    }

    const actualField = input[field];
    const actualType = typeof actualField;

    if (fieldType.constructor.name === 'GraphQLNonNull') {
      if (actualField === null || actualField === undefined) {
        validationErrors.push(`${field} is required.`);
        return validationErrors.length ? validationErrors : [];
      }

      fieldType = fieldType.ofType;
    }

    if (!actualField) {
      continue;
    }

    switch (fieldType.constructor.name) {
      case 'GraphQLNonNull':
        break;
      case 'GraphQLList':
        if (!actualField) {
          break;
        }

        if (!Array.isArray(actualField)) {
          validationErrors.push(
            `${field} is required of type array but provided ${typeof input[
              field
            ]}`
          );

          break;
        }

        {
          const re = actualField
            .map((i: any) => {
              let ft = fieldType.ofType;
              if (ft.constructor.name === 'GraphQLNonNull') {
                ft = ft.ofType;
              }
              return validateInput(i, ft);
            })
            .reduce((acc: string[], errors: string[], index: number) => {
              return [
                ...acc,
                ...errors.map((i) => {
                  return `inside ${field}[${index}] ${i}`;
                }),
              ];
            }, []);

          validationErrors = [...validationErrors, ...re];
        }

        break;
      case 'GraphQLScalarType':
        switch (`${fieldType}`) {
          case 'Int':
            if (actualType !== 'number') {
              validationErrors = [
                ...validationErrors,
                `field ${field} must be an integer but provided ${actualType}`,
              ];
            }
            break;
          case 'Float':
            if (actualType !== 'number') {
              validationErrors = [
                ...validationErrors,
                `field ${field} must be a float but provided ${actualType}`,
              ];
            }
            break;
          case 'String':
            if (actualType !== 'string') {
              validationErrors = [
                ...validationErrors,
                `field ${field} must be a string but provided ${actualType}`,
              ];
            }
            break;
          case 'Map':
            if (actualField instanceof Object) {
              break;
            }

            validationErrors = [
              ...validationErrors,
              `field ${field} must be a Object but provided ${actualType}`,
            ];
            break;
          default:
            console.log(
              `unknown field type ${fieldType} ${JSON.stringify(
                actualField,
                null,
                2
              )}`
            );
        }

        break;
      case 'GraphQLInputObjectType':
        if (!actualField) {
          break;
        }
        if (actualField instanceof Object) {
          break;
        }

        validationErrors = [
          ...validationErrors,
          `field ${field} must be a Object but provided ${actualType}`,
        ];
        break;

        break;
      default:
        console.log(
          `Unknown field type ${fieldType.constructor.name} ${JSON.stringify(
            input[field],
            null,
            2
          )}`
        );
    }

    if (actualField) {
      switch (fieldType.name) {
        case 'String':
          if (actualType !== 'string')
            validationErrors.push(
              `${field} should be a string but provided ${actualType}.`
            );
          break;

        case 'Boolean':
          if (actualType !== 'boolean')
            validationErrors.push(
              `${field} should be a boolean but provided ${actualType}.`
            );
          console.log('Boolean', actualField);
          break;

        case 'Map':
          if (actualType !== 'object')
            validationErrors.push(`${field} should be a object ${actualType}.`);
          break;

        default:
          validationErrors = [
            ...validationErrors,
            ...validateInput(
              actualField,
              schema.getType(fieldType.name || fieldType.ofType.name)
            ).map((i) => {
              return `inside ${field} ${i}`;
            }),
          ];
      }
    }
  }

  return validationErrors.length ? validationErrors : [];
}

export const validateType = (data: any, inputType: string) => {
  return validateInput(data, schema.getType(inputType));
};

export const useValidateType = (
  data: any,
  inputType: string,
  dep: any[] = []
) =>
  useMemo(() => {
    if (typeof window === 'undefined' || !data) {
      return [];
    }
    try {
      const res = validateInput(data, schema.getType(inputType));

      return res;
    } catch (e) {
      const er = e as Error;
      return [er.message];
    }
  }, dep);

const _test = () => {
  const accountIn = {
    contactEmail: 'sample@gmail.com',
    displayName: 'hi',
    metadata: {
      name: 'hi',
    },
    spec: {
      containers: [{ hi: 'hello' }],
    },
  };

  const result = validateType(accountIn, 'AppIn');
  console.log(result);
};
