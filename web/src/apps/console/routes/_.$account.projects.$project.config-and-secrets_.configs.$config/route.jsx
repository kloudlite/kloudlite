import { Plus, PlusFill } from '@jengaicons/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import Wrapper from '~/console/components/wrapper';
import { dummyData } from '~/console/dummy/data';
import ResourceList from '~/console/components/resource-list';
import Tools from './tools';
import Resources from './resources';

const Config = () => {
  const [data, setData] = useState(dummyData.configs);
  console.log(data);
  return (
    <Wrapper
      header={{
        title: 'kloud-root-ca.crt',
        backurl: '../configs',
        action: data.length > 0 && (
          <Button variant="basic" content="Add new entry" prefix={PlusFill} />
        ),
      }}
      empty={{
        is: data.length === 0,
        title: 'This is where you’ll manage your projects.',
        content: (
          <p>You can create a new project and manage the listed project.</p>
        ),
        action: {
          content: 'Add new entry',
          prefix: Plus,
        },
      }}
    >
      <div className="flex flex-col">
        <Tools />
      </div>
      <ResourceList>
        {data.map((d) => (
          <ResourceList.ResourceItem key={d.key} textValue={d.key}>
            <Resources item={d} />
          </ResourceList.ResourceItem>
        ))}
      </ResourceList>
    </Wrapper>
  );
};

export default Config;
