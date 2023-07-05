import * as TB from '@radix-ui/react-toolbar';
import ToggleGroup from './togglegroup';


export default Toolbar = ({ }) => {
    return (
        <TB.Root>
            <ToggleGroup value='left'>
                <ToggleGroup.Button content={"Left"} value={"left"} />
                <ToggleGroup.Button content={"Left"} value={"left"} />
            </ToggleGroup>
            <ToggleGroup value='left'>
                <ToggleGroup.Button content={"Left"} value={"left"} />
                <ToggleGroup.Button content={"Left"} value={"left"} />
            </ToggleGroup>
        </TB.Root>
    )
}