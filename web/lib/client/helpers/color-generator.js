const colorCode = (str = 'Sample') => {
  let hash = 0;
  for (let i = 0; i < str.length; i += 1) {
    // eslint-disable-next-line no-bitwise
    hash = str.charCodeAt(i) + ((hash << 5) - hash);
    // hash += str.charCodeAt(i);
  }
  let color = '#';
  for (let i = 0; i < 3; i += 1) {
    // eslint-disable-next-line no-bitwise
    const value = (hash >> (i * 8)) & 0xff;
    color += `00${value.toString(16)}`.substr(-2);
  }
  return color;
};

export const bgImage = (str = '#') => {
  return `linear-gradient(45deg, ${colorCode(str)}, ${colorCode(`${str}1`)})`;
};

export default colorCode;
