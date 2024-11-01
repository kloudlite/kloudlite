import relativeTime from 'dayjs/plugin/relativeTime';
import _dayjs from 'dayjs';
import localizedFormat from 'dayjs/plugin/localizedFormat';

_dayjs.extend(relativeTime);
_dayjs.extend(localizedFormat);

export const dayjs = _dayjs;
