const getQueries = (ctx: { params?: { provider: any }; request?: any }) => {
  const url = new URL(ctx.request.url);
  // logger.log(url.searchParams);
  return Object.fromEntries(url.searchParams.entries());
};

export default getQueries;
