/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VERSION: string;
  readonly VITE_CHANGELOG: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
