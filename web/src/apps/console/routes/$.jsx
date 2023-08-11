import { Link } from '@remix-run/react';

const _404 = () => {
  return (
    <div className="text-[5vw] flex gap-[1vw] justify-center items-center min-h-screen">
      <div className="flex flex-col items-center">
        <span className="text-text-critical text-[10vw]">404</span>
        <span className="text-text-warning uppercase animate-pulse">
          page not found
        </span>
        <Link to="/">
          <a className="text-text-primary text-[1rem] hover:underline hover:text-text-strong transition-all underline">
            Home Page
          </a>
        </Link>
      </div>
    </div>
  );
};

export const meta = () => {
  return [{ title: '404 | Not found' }];
};

export default _404;
