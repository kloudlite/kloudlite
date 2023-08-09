import { MinusCircle } from '@jengaicons/react';
import { useEffect, useState } from 'react';
import { Button, IconButton } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import { uuid } from '~/components/utils';
import * as Popup from '~/components/molecule/popup';
import * as SelectInput from '~/components/atoms/select';
import { dummyData } from '~/console/dummy/data';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';

const Labels = ({ onChange, value }) => {
  const newItem = [{ key: '', value: '', id: uuid() }];
  const [items, setItems] = useState(newItem);

  const handleChange = (_value, id, target) => {
    setItems(
      items.map((i) => {
        if (i.id === id) {
          switch (target) {
            case 'key':
              return { ...i, key: _value };
            case 'value':
            default:
              return { ...i, value: _value };
          }
        }
        return i;
      })
    );
  };

  useEffect(() => {
    if (onChange) onChange(Array.from(items));
  }, [items]);

  useEffect(() => {
    if (value?.length > 0) {
      setItems(Array.from(value));
    }
  }, []);

  return (
    <div className="flex flex-col gap-xl">
      <div className="flex flex-col gap-md">
        <span className="text-text-default bodyMd-medium">Labels</span>
        {items.map((item) => (
          <div key={item.id} className="flex flex-row gap-xl items-end">
            <div className="flex-1">
              <TextInput
                placeholder="Key"
                value={item.key}
                onChange={({ target }) =>
                  handleChange(target.value, item.id, 'key')
                }
              />
            </div>
            <div className="flex-1">
              <TextInput
                placeholder="Value"
                value={item.value}
                onChange={({ target }) =>
                  handleChange(target.value, item.id, 'value')
                }
              />
            </div>
            <IconButton
              icon={MinusCircle}
              variant="plain"
              disabled={items.length < 2}
              onClick={() => {
                setItems(items.filter((i) => i.id !== item.id));
              }}
            />
          </div>
        ))}
      </div>
      <Button
        variant="basic"
        content="Add label"
        size="sm"
        onClick={() => {
          setItems([...items, { ...newItem[0], id: uuid() }]);
        }}
      />
    </div>
  );
};

const Taints = ({ onChange, value }) => {
  const newItem = { taint: '', type: '', value: '', id: uuid() };
  const [items, setItems] = useState([newItem]);
  const [taints, _setTaints] = useState(dummyData.taints);

  const handleChange = (_value, id, target) => {
    setItems(
      items.map((i) => {
        if (i.id === id) {
          switch (target) {
            case 'type':
              return { ...i, type: _value };
            case 'value':
              return { ...i, value: _value };
            case 'taint':
            default:
              return { ...i, taint: _value };
          }
        }
        return i;
      })
    );
  };

  useEffect(() => {
    if (onChange) onChange(Array.from(items));
  }, [items]);

  useEffect(() => {
    if (value?.length > 0) {
      setItems(Array.from(value));
    }
  }, []);

  return (
    <div className="flex flex-col gap-xl">
      <div className="flex flex-col gap-md">
        <span className="text-text-default bodyMd-medium">Taints</span>
        {items.map((item) => (
          <div key={item.id} className="flex flex-row gap-xl items-end">
            <div className="flex-1">
              <SelectInput.Select
                value={item.taint}
                onChange={(e) => {
                  handleChange(e, item.id, 'taint');
                }}
              >
                <SelectInput.Option>--Select--</SelectInput.Option>
                {taints.map((tts) => (
                  <SelectInput.Option key={tts.id} value={tts.value}>
                    {tts.label}
                  </SelectInput.Option>
                ))}
              </SelectInput.Select>
            </div>
            <div className="flex-1">
              <TextInput
                placeholder="Type"
                value={item.type}
                onChange={({ target }) =>
                  handleChange(target.value, item.id, 'type')
                }
              />
            </div>
            <div className="flex-1">
              <TextInput
                placeholder="Value"
                value={item.value}
                onChange={({ target }) =>
                  handleChange(target.value, item.id, 'value')
                }
              />
            </div>
            <IconButton
              icon={MinusCircle}
              variant="plain"
              disabled={items.length < 2}
              onClick={() => {
                setItems(items.filter((i) => i.id !== item.id));
              }}
            />
          </div>
        ))}
      </div>
      <Button
        variant="basic"
        content="Add taint"
        size="sm"
        onClick={() => {
          setItems([...items, { ...newItem, id: uuid() }]);
        }}
      />
    </div>
  );
};

const HandleNodePool = ({ show, setShow }) => {
  const [nodePlans, _setNodePlans] = useState(dummyData.nodePlans);
  const [provisionTypes, _setProvisionTypes] = useState(
    dummyData.provisionTypes
  );

  const { values, errors, handleChange, handleSubmit, resetValues } = useForm({
    initialValues: {
      name: '',
      minimum: '',
      maximum: '',
      nodeplan: '',
      provisiontype: '',
      labels: [],
      taints: [],
    },
    validationSchema: Yup.object({}),
    onSubmit: () => {},
  });

  return (
    <Popup.PopupRoot
      show={show}
      onOpenChange={(e) => {
        if (!e) {
          resetValues();
        }

        setShow(e);
      }}
    >
      <Popup.Header>
        {show?.type === 'add' ? 'Add nodepool' : 'Edit nodepool'}
      </Popup.Header>
      <form onSubmit={handleSubmit}>
        <Popup.Content>
          <div className="flex flex-col gap-2xl">
            <TextInput
              label="Name"
              value={values.name}
              onChange={handleChange('name')}
            />
            <div className="flex flex-row gap-xl items-end">
              <div className="flex-1">
                <TextInput
                  label="Capacity"
                  placeholder="Minimum"
                  value={values.minimum}
                  onChange={handleChange('minimum')}
                />
              </div>
              <div className="flex-1">
                <TextInput
                  placeholder="Maximum"
                  value={values.maximum}
                  onChange={handleChange('maximum')}
                />
              </div>
            </div>
            <SelectInput.Select
              value={values.nodeplan}
              label="Node plan"
              onChange={(value) =>
                handleChange('nodeplan')({ target: { value } })
              }
            >
              <SelectInput.Option disabled value="">
                --Select--
              </SelectInput.Option>
              {nodePlans.map((nodeplan) => (
                <SelectInput.Option
                  key={nodeplan.id}
                  disabled={nodeplan.disabled}
                  value={nodeplan.value}
                >
                  {nodeplan.label}
                </SelectInput.Option>
              ))}
            </SelectInput.Select>

            {show?.type === 'add' && (
              <SelectInput.Select
                value={values.provisiontype}
                label="Provision type"
                onChange={(value) =>
                  handleChange('provisiontype')({ target: { value } })
                }
              >
                <SelectInput.Option disabled value="">
                  --Select--
                </SelectInput.Option>
                {provisionTypes.map((pt) => (
                  <SelectInput.Option value={pt.value} key={pt.id}>
                    {pt.label}
                  </SelectInput.Option>
                ))}
              </SelectInput.Select>
            )}

            <Labels
              value={values.labels}
              onChange={(value) =>
                handleChange('labels')({ target: { value } })
              }
            />
            <Taints
              onChange={(value) =>
                handleChange('taints')({ target: { value } })
              }
            />
          </div>
        </Popup.Content>
        <Popup.Footer>
          <Popup.Button closable content="Cancel" variant="basic" />
          <Popup.Button type="submit" content="Save" variant="primary" />
        </Popup.Footer>
      </form>
    </Popup.PopupRoot>
  );
};

export default HandleNodePool;
