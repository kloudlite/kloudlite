/* eslint-disable no-bitwise */
type Contrast = 'light' | 'dark';

function generateColorFromName(
  name: string,
  targetColor: string,
  contrast: Contrast
): string {
  // Convert name to hash
  let hash = 0;

  for (let i = 0; i < name.length; i += 1) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash);
  }

  // Convert hash to RGB
  const r = (hash & 0xff0000) >> 16;
  const g = (hash & 0x00ff00) >> 8;
  const b = hash & 0x0000ff;

  // Parse target color
  const targetR = parseInt(targetColor.slice(1, 3), 16);
  const targetG = parseInt(targetColor.slice(3, 5), 16);
  const targetB = parseInt(targetColor.slice(5, 7), 16);

  // Adjust color to be close to target color
  const adjustedR = Math.round((r + targetR) / 2);
  const adjustedG = Math.round((g + targetG) / 2);
  const adjustedB = Math.round((b + targetB) / 2);

  // Adjust brightness
  function adjustBrightness(color: number, factor: number): number {
    return Math.min(255, Math.max(0, color + factor));
  }

  const factor = contrast === 'light' ? 50 : -50;
  const finalR = adjustBrightness(adjustedR, factor);
  const finalG = adjustBrightness(adjustedG, factor);
  const finalB = adjustBrightness(adjustedB, factor);

  // Convert RGB to hex
  return `#${finalR.toString(16).padStart(2, '0')}${finalG
    .toString(16)
    .padStart(2, '0')}${finalB.toString(16).padStart(2, '0')}`;
}

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

const generateColor = (str = '#', contrast: Contrast = 'dark') => {
  const cc = colorCode(str);
  return `linear-gradient(45deg, ${generateColorFromName(
    str,
    cc,
    contrast
  )}, ${generateColorFromName(`${str + str + str}`, cc, contrast)})`;
};

export default generateColor;
