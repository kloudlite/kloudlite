const _404 = () => {
  return (
    <div className="text-[10rem] flex justify-center items-center min-h-screen text-text-critical animate-pulse">
      404 Not found
    </div>
  );
};

export const meta = () => {
  return [{ title: '404 | Not found' }];
};

export default _404;
