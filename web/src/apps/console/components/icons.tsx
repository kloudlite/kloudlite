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

export const EnvIconComponent = ({ size = 16 }: { size?: number }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 20 20"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    <path
      fillRule="evenodd"
      clipRule="evenodd"
      d="M5.97709 1.69811C5.82539 1.61322 5.64048 1.61322 5.48877 1.69811L0.381816 4.55584C0.223834 4.64424 0.125977 4.81114 0.125977 4.99217V10.7062C0.125977 10.8872 0.223834 11.0541 0.381816 11.1425L5.48877 14.0002C5.64048 14.0851 5.82539 14.0851 5.97709 14.0002L10.3343 11.5621V15.4764C10.3343 15.6574 10.4322 15.8243 10.5901 15.9127L14.8597 18.3019C15.0115 18.3868 15.1964 18.3868 15.3481 18.3019L19.6177 15.9127C19.7757 15.8243 19.8735 15.6574 19.8735 15.4764V10.709C19.8745 10.6597 19.8682 10.6099 19.8541 10.5612C19.8182 10.4362 19.7343 10.3282 19.6177 10.2629L15.3481 7.87376C15.1964 7.78887 15.0115 7.78887 14.8597 7.87376L11.3399 9.84339V5.00472C11.3413 4.94648 11.3326 4.88749 11.313 4.83046C11.2737 4.71558 11.1933 4.61695 11.084 4.55584L5.97709 1.69811ZM9.8223 4.99571L5.73293 2.7074L1.68137 4.97455L5.78948 7.26275H5.7897C5.80297 7.26275 5.81612 7.26327 5.82913 7.26428L9.8223 4.99571ZM1.12598 5.80986L5.2897 8.12903V12.7429L1.12598 10.413V5.80986ZM10.3399 10.413L6.2897 12.6794V8.15274L10.3399 5.85178V10.413ZM18.3559 10.7028L15.1039 8.88305L11.8848 10.6844L15.1686 12.5135L18.3559 10.7028ZM11.3343 11.5224V15.1832L14.6515 17.0395V13.3701L11.3343 11.5224ZM18.8735 15.1832L15.6515 16.9862V13.3892L18.8735 11.5588V15.1832Z"
      fill="#111827"
    />
  </svg>
);

export const EnvTemplateIconComponent = ({ size = 16 }: { size?: number }) => (
  <svg
    width={16}
    height={12}
    viewBox="0 0 16 12"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    <path
      d="M4.37097 0.677429C0.5 0.677429 4.37097 6.00001 0.5 6.00001C4.37097 6.00001 0.5 11.3226 4.37097 11.3226"
      stroke="#111827"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      d="M11.6289 0.677429C15.4999 0.677429 11.6289 6.00001 15.4999 6.00001C11.6289 6.00001 15.4999 11.3226 11.6289 11.3226"
      stroke="#111827"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
);

export {
  AWSlogoFill,
  ArrowClockwise,
  ArrowCounterClockwise,
  ArrowFatDown as ArrowDown,
  ArrowLeftLgFill as ArrowLeft,
  ArrowLineDown,
  ArrowRightLgFill as ArrowRight,
  ArrowFatUp as ArrowUp,
  ArrowsClockwise,
  ArrowsCounterClockwise,
  ArrowsDownUp,
  BackingServices,
  BellFill,
  Buildings,
  CalendarCheckFill,
  CaretDownFill,
  Check,
  CheckCircle,
  CheckCircleFill,
  ChecksFill as Checks,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  ChevronUp,
  ChevronUpDown,
  Circle,
  CircleFill,
  CircleNotch,
  CircleWavyCheckFill,
  CirclesFour,
  Clock,
  CodeSimpleFill,
  Container,
  Copy,
  CopySimple,
  Cpu,
  Crosshair,
  Database,
  Dockerlogo,
  Domain,
  DotsSix,
  DotsThreeOutlineFill,
  DotsThreeVerticalFill,
  DownloadSimple,
  Eye,
  EyeSlash,
  Fan,
  File,
  FileLock,
  GearFill,
  GearSix,
  GitBranch,
  GitBranchFill,
  GitMerge,
  GithubLogoFill,
  GitlabLogoFill,
  Globe,
  GoogleCloudlogo,
  HamburgerFill,
  HouseLine,
  Info,
  InfoFill,
  InfraAsCode,
  Link,
  LinkBreak,
  List,
  ListBullets,
  ListDashes,
  ListNumbers,
  LockSimple,
  LockSimpleOpen,
  MinusCircle,
  NoOps,
  Nodeless,
  Note,
  Pause,
  PencilLine as Pencil,
  PencilLine,
  PencilSimple,
  Play,
  PlugsConnected,
  Plus,
  Project,
  QrCode,
  Question,
  Repeat,
  Search,
  ShieldCheck,
  SignOut,
  Sliders,
  Smiley,
  SmileySad,
  Spinner,
  SquaresFour,
  StackSimple,
  Tag,
  TerminalWindow,
  Trash,
  TreeStructure,
  UserCircle,
  Users,
  VirtualMachine,
  Warning,
  WarningCircle,
  WarningCircleFill,
  WarningFill,
  WarningOctagonFill,
  WireGuardlogo,
  X,
  XCircleFill,
  XFill
} from '@jengaicons/react';

