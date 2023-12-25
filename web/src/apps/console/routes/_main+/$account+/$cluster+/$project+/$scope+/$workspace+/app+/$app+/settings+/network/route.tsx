import { TitleBox } from '~/console/components/raw-wrapper';
import { FadeIn } from '~/console/page-components/util';
import { ExposedPorts } from '../../../../new-app/app-network';



const AppNetwork = () => {
  return (
    <FadeIn>
      <TitleBox
        title="Network"
        subtitle="Expose service ports that need to be exposed from container"
      />
      <ExposedPorts />
    </FadeIn>
  );
};

export default AppNetwork;
