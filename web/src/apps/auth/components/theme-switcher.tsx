import { Monitor, Moon, Sun } from '@jengaicons/react';
import ButtonGroup from '~/components/atoms/button-group';
import { useTheme } from '~/root/lib/client/hooks/useTheme';

const ThemeSwitcher = () => {
  const { theme, setTheme } = useTheme();
  return (
    <ButtonGroup.Root
      variant="outline"
      selectable
      value={theme}
      onValueChange={(v: any) => {
        setTheme(v);
      }}
    >
      <ButtonGroup.IconButton value="light" icon={<Sun />} />
      <ButtonGroup.IconButton value="dark" icon={<Moon />} />
      <ButtonGroup.IconButton value="system" icon={<Monitor />} />
    </ButtonGroup.Root>
  );
};

export default ThemeSwitcher;
