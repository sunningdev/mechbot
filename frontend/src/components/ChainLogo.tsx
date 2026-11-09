type Props = { chain: string; size?: number };

export default function ChainLogo({ chain, size = 24 }: Props) {
  const common = { width: size, height: size, viewBox: "0 0 32 32", "aria-hidden": true } as const;
  switch (chain) {
    case "ethereum":
      return (
        <svg {...common}>
          <circle cx="16" cy="16" r="16" fill="#627EEA" />
          <path d="M16 4.5v8.5l7.2 3.2z" fill="#fff" fillOpacity=".6" />
          <path d="M16 4.5L8.8 16.2 16 13z" fill="#fff" />
          <path d="M16 21.7v5.8l7.2-10z" fill="#fff" fillOpacity=".6" />
          <path d="M16 27.5v-5.8l-7.2-4.2z" fill="#fff" />
          <path d="M16 20.4l7.2-4.2L16 13z" fill="#fff" fillOpacity=".2" />
          <path d="M8.8 16.2l7.2 4.2V13z" fill="#fff" fillOpacity=".6" />
        </svg>
      );
    case "bsc":
      return (
        <svg {...common}>
          <circle cx="16" cy="16" r="16" fill="#F0B90B" />
          <g fill="#fff">
            <path d="M12.1 14.2L16 10.3l3.9 3.9 2.3-2.3L16 5.7l-6.2 6.2z" />
            <path d="M5.7 16l2.3-2.3 2.3 2.3-2.3 2.3z" />
            <path d="M12.1 17.8L16 21.7l3.9-3.9 2.3 2.3L16 26.3l-6.2-6.2z" />
            <path d="M21.7 16l2.3-2.3 2.3 2.3-2.3 2.3z" />
            <path d="M18.3 16L16 13.7 14.3 15.4l-.2.2-.4.4L16 18.3z" />
          </g>
        </svg>
      );
    case "arbitrum":
      return (
        <svg {...common}>
          <circle cx="16" cy="16" r="16" fill="#213147" />
          <path d="M16 5.5l9.1 5.25v10.5L16 26.5l-9.1-5.25v-10.5z" fill="none" stroke="#9DCCED" strokeWidth="1.4" />
          <path d="M17.4 11.6l4.6 7.9-1.9 1.1-4.1-7z" fill="#12AAFF" />
          <path d="M14.6 11.6l-4.6 7.9 1.9 1.1 4.1-7z" fill="#fff" />
          <path d="M13 22.6l3-5.2 3 5.2-1.4.8L16 20.6l-1.6 2.8z" fill="#12AAFF" />
        </svg>
      );
    case "base":
      return (
        <svg {...common}>
          <circle cx="16" cy="16" r="16" fill="#0052FF" />
          <rect x="9" y="9" width="14" height="14" rx="2.5" fill="#fff" />
        </svg>
      );
    case "optimism":
      return (
        <svg {...common}>
          <circle cx="16" cy="16" r="16" fill="#FF0420" />
          <text x="16" y="20.5" textAnchor="middle" fontFamily="Arial, sans-serif" fontWeight="700" fontStyle="italic" fontSize="12" fill="#fff">
            OP
          </text>
        </svg>
      );
    case "tron":
      return (
        <svg {...common}>
          <circle cx="16" cy="16" r="16" fill="#EF0027" />
          <path
            d="M21.9 9.3L6.4 6.4l8.2 20.6 11.4-13.8zm-.3 1.6l2.4 2.3-6.5 1.2zM16 14.1L9.2 8.4l11 2zm-1 .2l-1.1 9.1L8 8.9zm1 .5l7.1-1.3-8.1 9.8z"
            fill="#fff"
          />
        </svg>
      );
    default:
      return (
        <svg {...common}>
          <circle cx="16" cy="16" r="16" fill="#999" />
        </svg>
      );
  }
}
