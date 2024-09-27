import React from 'react';

export const AvatarNotification = ({ size = 16 }: { size?: number }) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width={size}
      height={size}
      fill="none"
      viewBox="0 0 25 26"
    >
      <rect width="24" height="24" x="0.5" y="1" fill="#fff" rx="12" />
      <rect width="24" height="24" x="0.5" y="1" stroke="#E4E4E7" rx="12" />
      <path
        fill="#111827"
        d="M4.813 13l3.17-8.73H9.27L12.482 13h-1.171L9.458 7.82c-.113-.32-.246-.718-.398-1.195-.149-.48-.33-1.092-.545-1.834h.21c-.21.75-.394 1.37-.55 1.857-.153.485-.28.875-.381 1.172L5.996 13H4.812zm1.593-2.438V9.59h4.483v.973H6.406zm10.237 2.59c-.63 0-1.176-.101-1.641-.304-.461-.204-.822-.487-1.084-.85a2.468 2.468 0 01-.445-1.283h1.142c.031.328.143.6.334.814.192.211.436.37.733.475.3.101.62.152.96.152a2.76 2.76 0 001.061-.193c.317-.129.567-.31.75-.545.184-.234.276-.506.276-.814 0-.282-.08-.51-.24-.686a1.756 1.756 0 00-.628-.434A6.52 6.52 0 0017 9.186l-1.055-.3c-.703-.198-1.25-.484-1.64-.855-.39-.37-.586-.845-.586-1.424 0-.492.133-.921.398-1.289a2.602 2.602 0 011.078-.861c.453-.203.961-.305 1.524-.305.574 0 1.082.102 1.523.305.442.203.79.48 1.043.832.258.348.395.742.41 1.184h-1.09a1.302 1.302 0 00-.609-.985c-.355-.234-.793-.351-1.312-.351-.375 0-.706.062-.99.187-.282.121-.5.29-.657.504-.156.211-.234.453-.234.727 0 .304.093.55.281.738.191.184.416.328.674.434.262.101.498.181.709.24l.873.24c.234.063.492.148.773.258.285.11.557.256.815.44.258.179.468.41.633.69.168.278.252.62.252 1.026 0 .477-.125.906-.375 1.29-.247.382-.606.685-1.079.907-.472.223-1.045.334-1.716.334z"
      />
      <rect width="24" height="24" x="0.5" y="1" fill="#fff" rx="12" />
      <path
        fill="#4B5563"
        fillRule="evenodd"
        d="M12.5 11.786c1.974 0 3.574-1.63 3.574-3.64 0-2.012-1.6-3.642-3.574-3.642-1.974 0-3.574 1.63-3.574 3.641s1.6 3.641 3.574 3.641zm0 9.71c3.003 0 5.686-1.423 7.434-3.641-1.748-2.219-4.431-3.641-7.434-3.641s-5.686 1.422-7.434 3.64c1.748 2.22 4.431 3.642 7.434 3.642z"
        clipRule="evenodd"
      />
      <circle cx="4.5" cy="5" r="4" fill="#DC2626" />
    </svg>
  );
};

export {
  Warning,
  WarningCircleFill,
  Domain,
  ArrowLeftLgFill as ArrowLeft,
  ArrowRightLgFill as ArrowRight,
  ArrowFatUp as ArrowUp,
  ArrowFatDown as ArrowDown,
  ArrowsDownUp,
  Plus,
  Trash,
  PencilLine as Pencil,
  PencilSimple,
  GithubLogoFill,
  GitlabLogoFill,
  GitBranchFill,
  Users,
  Check,
  ChevronLeft,
  ChevronRight,
  X,
  SmileySad,
  InfoFill,
  CheckCircleFill,
  WarningFill,
  WarningOctagonFill,
  LockSimple,
  XCircleFill,
  LockSimpleOpen,
  MinusCircle,
  Search,
  ArrowsCounterClockwise,
  ArrowClockwise,
  Copy,
  GearSix,
  QrCode,
  WireGuardlogo,
  ChevronUpDown,
  ChevronDown,
  Buildings,
  Project,
  InfraAsCode,
  Container,
  File,
  TreeStructure,
  CirclesFour,
  BackingServices,
  VirtualMachine,
  Database,
  ArrowsClockwise,
  Info,
  Fan,
  WarningCircle,
  ChecksFill as Checks,
  CircleNotch,
  Circle,
  CircleFill,
  Spinner,
  Globe,
  ShieldCheck,
  NoOps,
  Nodeless,
  GitMerge,
  PencilLine,
  AWSlogoFill,
  GoogleCloudlogo,
  ArrowCounterClockwise,
  CopySimple,
  DotsThreeVerticalFill,
  CaretDownFill,
  Question,
  ListBullets,
  ChevronUp,
  List,
  SquaresFour,
  Smiley,
  ArrowLineDown,
  Clock,
  ListNumbers,
  DotsSix,
  SignOut,
  Repeat,
  LinkBreak,
  Link,
  Eye,
  Cpu,
  Crosshair,
  HouseLine,
  CircleWavyCheckFill,
  DownloadSimple,
  GitBranch,
  Tag,
  CodeSimpleFill,
  TerminalWindow,
  DotsThreeOutlineFill,
  GearFill,
  HamburgerFill,
  XFill,
  EyeSlash,
  CheckCircle,
  CalendarCheckFill,
  PlugsConnected,
  ListDashes,
  BellFill,
  UserCircle,
  Sliders,
  FileLock,
  StackSimple,
  Play,
  Pause,
  Dockerlogo,
} from '@jengaicons/react';
