import { createCookieSessionStorage } from '@remix-run/node';

const { getSession, commitSession, destroySession } =
  createCookieSessionStorage({
    // a Cookie from `createCookie` or the CookieOptions to create one
    cookie: {
      // firebase token
      name: 'user-configs',

      // all of these are optional
      expires: new Date(Date.now() + 600),
      httpOnly: true,
      maxAge: 600,
      path: '/',
      sameSite: 'lax',
      secrets: ['tacos'],
      secure: true,
    },
  });

export { getSession, commitSession, destroySession };
