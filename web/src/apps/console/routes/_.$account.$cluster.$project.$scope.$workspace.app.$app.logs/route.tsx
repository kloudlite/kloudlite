import { defer } from '@remix-run/node';
import { useLoaderData } from '@remix-run/react';
import { mapper } from '~/components/utils';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import HighlightJsLog from '~/console/components/logger';

// const Item = ({
//   item,
//   searchText,
// }: {
//   searchText: string;
//   item: { id: string; title: string };
// }) => {
//   return (
//     <div
//       key={item.id}
//       className={classNames('item overflow-hidden', {
//         'bg-text-critical': item.title.indexOf(searchText) !== -1,
//       })}
//       style={{ height: 20 }}
//     >
//       {item.id}.{item.title}
//     </div>
//   );
// };
//
// const Logger = ({ items }: { items: { id: string; title: string }[] }) => {
//   const ref = useRef<HTMLDivElement | null>(null);
//   const [searchText, setSearchText] = useState('');
//   return (
//     <div className="flex flex-col gap-2xl">
//       <TextInput
//         onChange={(e) => setSearchText(e.target.value)}
//         value={searchText}
//       />
//       <div
//         className="scroll-container max-h-[30vh] overflow-auto border"
//         ref={ref}
//       >
//         <ViewportList
//           viewportRef={ref}
//           items={items.filter((i) => i.title.indexOf(searchText) !== -1)}
//           itemSize={20}
//         >
//           {(item) => <Item item={item} searchText={searchText} />}
//         </ViewportList>
//       </div>
//     </div>
//   );
// };

export const loader = () => {
  const promise = pWrapper(async () => {
    const items: string[] = mapper(
      // @ts-ignore
      new Array(10000).fill(),
      (_: any, index) => {
        return `this is ${index + 1}th item`;
      }
    );
    return { items };
  });
  return defer({ promise });
};

const ItemList = () => {
  const { promise } = useLoaderData<typeof loader>();

  return (
    <LoadingComp data={promise}>
      {({ items }) => {
        return <HighlightJsLog dark text={items.join('\n')} />;
      }}
    </LoadingComp>
  );
};

export default ItemList;
