import { Outlet } from '@remix-run/react';
import Container from '../pages/container';

const Console = () => {
  return (
    <Container>
      <Outlet />
    </Container>
  );
};

export default Console;
