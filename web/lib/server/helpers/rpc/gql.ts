// TODO: Not required for now, maybe we can use it later

// const withRPC = (handler, options) => {
//   return async (req, res, next) => {
//     try {
//       const method = req.body.method.split('.').reduce((acc, item) => {
//         return acc[item];
//       }, handler);
//       if (!method) throw new Error('handler Method not found');
//       const response = await method(...req.body.args);
//       res.json(response);
//     } catch (err) {
//       next(err);
//     }
//   };
// };

// export const handler = async ({ ctx, method, args, httpOptions }) => {
//   return ctx.post('/', {
//     method,
//     args,
//   });
// };
//
// export const makeGraphClient = (ctx, slack, options) => {
//   return new Proxy(() => {}, {
//     apply(target, method, args) {
//       return handler({
//         ctx,
//         method: slack,
//         args,
//       });
//     },
//     get(target, prefix) {
//       return makeGraphClient(ctx, slack ? [slack, prefix].join('.') : prefix);
//     },
//   });
// };
