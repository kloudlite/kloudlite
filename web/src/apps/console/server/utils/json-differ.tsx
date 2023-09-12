/* eslint-disable no-prototype-builtins */
/* eslint-disable no-restricted-syntax */

import * as yaml from 'js-yaml';
import React, { PureComponent } from 'react';
import ReactDiffViewer, { DiffMethod } from 'react-diff-viewer';

type JsonObject = { [key: string]: any };

function convertToYaml(diff: JsonObject): string {
  return yaml.dump(diff);
}

export class JsonDiffer {
  private static isObject(item: any): boolean {
    return item && typeof item === 'object' && !Array.isArray(item);
  }

  static diff(obj1: JsonObject, obj2: JsonObject): JsonObject {
    const diffResult: JsonObject = {};

    const recursiveDiff = (
      a: JsonObject,
      b: JsonObject,
      result: JsonObject
    ) => {
      for (const key in a) {
        if (!b.hasOwnProperty(key)) {
          result[key] = `Missing in second JSON: ${a[key]}`;
        } else if (this.isObject(a[key]) && this.isObject(b[key])) {
          const subDiff = {};
          recursiveDiff(a[key], b[key], subDiff);
          if (Object.keys(subDiff).length) {
            result[key] = subDiff;
          }
        } else if (a[key] !== b[key]) {
          result[key] = `Changed: ${a[key]} -> ${b[key]}`;
        }
      }

      for (const key in b) {
        if (!a.hasOwnProperty(key)) {
          result[key] = `Missing in first JSON: ${b[key]}`;
        }
      }
    };

    recursiveDiff(obj1, obj2, diffResult);
    return diffResult;
  }
}

const oldCode = `
{
  "name": "Original name",
  "description": null
}
`;
const newCode = `
{
  "name": "My updated name",
  "description": "Brand new description",
  "status": "running"
}
`;

export const DiffTest = ({ a, b }: { a: JsonObject; b: JsonObject }) => {
  return <ReactDiffViewer oldValue={oldCode} newValue={newCode} splitView />;
};

export default DiffTest;
