declare global {
  interface Window {
    grecaptcha: {
      enterprise: {
        ready: (callback: () => Promise<void>) => void;
        execute: (
          sitekey: string,
          action: {
            action: string;
          },
        ) => Promise<string>;
      };
    };
  }
}

export const onReady = (callback: () => Promise<void>) => {
  if (
    !window.grecaptcha ||
    !window.grecaptcha.enterprise ||
    !window.grecaptcha.enterprise.ready
  ) {
    console.warn('window.grecaptcha.enterprise.ready is not defined.');
    return;
  }
  const ready = window.grecaptcha.enterprise.ready;
  ready(callback);
};

export const execute = async (
  sitekey: string,
  action: {
    action: string;
  },
) => {
  if (
    !window.grecaptcha ||
    !window.grecaptcha.enterprise ||
    !window.grecaptcha.enterprise.ready
  ) {
    console.warn('window.grecaptcha.enterprises.ready is not defined.');
    return;
  }
  const exec = window.grecaptcha.enterprise.execute;
  return exec(sitekey, action);
};

const grecaptcha = {
  onReady,
  execute,
};

export default grecaptcha;
