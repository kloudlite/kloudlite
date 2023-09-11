/* eslint-disable no-restricted-syntax */
/* eslint-disable guard-for-in */
import { buildASTSchema, parse, GraphQLNamedType } from 'graphql';
import fs from 'fs';

const typeDefs = fs.readFileSync('./test/q.graphql');

const astDocument = parse(typeDefs.toString());
const schema = buildASTSchema(astDocument);

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

function validateInput(
  input: any,
  graphqlInputType: GraphQLNamedType | undefined
) {
  let validationErrors: string[] = [];

  // @ts-ignore
  const fields = graphqlInputType?._fields || [];

  for (const field in fields) {
    let fieldType = fields[field].type;

    // Check for Non-nullable fields
    if (fieldType.constructor.name === 'GraphQLNonNull') {
      if (input[field] === null || input[field] === undefined) {
        validationErrors.push(`${field} is required.`);
      }

      fieldType = fieldType.ofType;
    }

    if (
      fieldType.constructor.name === 'GraphQLList' &&
      typeof input[field] !== 'undefined'
    ) {
      if (!Array.isArray(input[field])) {
        validationErrors.push(
          `${field} is required of type array but found ${typeof input[field]}`
        );
      } else {
        const res = input[field]
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

        validationErrors = [...validationErrors, ...res];
      }
    }

    const actualType = typeof input[field];
    if (actualType !== 'undefined') {
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
              `${field} should be a boolean ${actualType}.`
            );
          break;

        case 'Map':
          if (actualType !== 'object')
            validationErrors.push(`${field} should be a object ${actualType}.`);
          break;

        default:
          // console.log(ft.constructor.name, ft.name, field);
          validationErrors = [
            ...validationErrors,
            ...validateInput(
              input[field],
              schema.getType(fieldType.name || fieldType.ofType.name)
            ).map((i) => {
              return `inside ${field} ${i}`;
            }),
          ];
        // validationErrors.push(`Not Supported ${ft.name}`);
      }
      if (fieldType.name === 'Boolean' && actualType !== 'boolean') {
        validationErrors.push(`${field} should be a boolean.`);
      }
    }
  }

  return validationErrors.length ? validationErrors : [];
}

const result = validateInput(accountIn, schema.getType('AppIn'));
console.log(result); // If all validations pass, it should print "Valid" or list of errors
