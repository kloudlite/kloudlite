import { Router } from 'express';
import { httpHandler } from '../../common/http-handler';
import testServices from '../services';

const router = Router();
const passport = require('passport');

const setCookie = (res, key = '', token = '', options = {}) => {
  const host = req.get('host');
  res.cookie('jwt', token, {
    ...{
      maxAge: 1000 * 60 * 60 * 24 * 7,
      httpOnly: true,
      secure: false,
      path: '/',
    },
    ...options,
    domain: host === 'localhost' ? '.abdhesh.com' : host,
  });
};

router.get('/set-cookie', (req, res) => {
  const origin = req.get('origin');

  setCookie(res, 'jwt', 'token');

  res.cookie('name', 'value', { domain: '.abdhesh.com' });

  res.send('cookie set');
});

export default router;
