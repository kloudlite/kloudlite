import { Link } from '@remix-run/react';

export const BaseLayout = ({ children }) => {
  return (
    <>
      <nav>
        <Link to="/pricing" prefetch>
          Pricing
        </Link>
        <Link to="/product" prefetch>
          Product
        </Link>
        <Link to="/features" prefetch>
          Features
        </Link>
        <Link to="/documentation" prefetch>
          Docs
        </Link>
        <Link to="/resources" prefetch>
          Resources
        </Link>
      </nav>
      {children}
    </>
  );
};
