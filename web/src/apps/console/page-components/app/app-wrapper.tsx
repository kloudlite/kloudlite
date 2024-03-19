import { ReactNode } from 'react';
import { Button } from '~/components/atoms/button';
import Wrapper from '~/console/components/wrapper';
import { useUnsavedChanges } from '~/root/lib/client/hooks/use-unsaved-changes';

const AppWrapper = ({
  children,
  title,
}: {
  children: ReactNode;
  title: string;
}) => {
  const { setPerformAction, hasChanges, loading } = useUnsavedChanges();
  return (
    <div>
      <Wrapper
        secondaryHeader={{
          title,
          action: hasChanges && !loading && (
            <div className="flex flex-row items-center gap-lg">
              <Button
                disabled={loading}
                variant="basic"
                content="Discard changes"
                onClick={() => setPerformAction('discard-changes')}
              />
              <Button
                disabled={loading}
                content={loading ? 'Committing changes.' : 'View changes'}
                loading={loading}
                onClick={() => setPerformAction('view-changes')}
              />
            </div>
          ),
        }}
      >
        {children}
      </Wrapper>
    </div>
  );
};
export default AppWrapper;
