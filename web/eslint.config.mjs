// web/eslint.config.mjs
import js from '@eslint/js';
import ts from 'typescript-eslint';
import next from '@next/eslint-plugin-next';
import importPlugin from 'eslint-plugin-import';
import unused from 'eslint-plugin-unused-imports';

export default [
  // 無視リスト（フレームワーク/生成物/型定義を除外）
  { ignores: ['.next/**', 'node_modules/**', 'dist/**', 'coverage/**', '**/*.d.ts'] },

  js.configs.recommended,
  ...ts.configs.recommended,

  {
    files: ['app/**/*.{ts,tsx,js,jsx}'],

    plugins: {
      '@next/next': next,
      import: importPlugin,
      'unused-imports': unused,
    },
    rules: {
      // Next 推奨を適用
      ...next.configs.recommended.rules,
      ...next.configs['core-web-vitals'].rules,

      'unused-imports/no-unused-imports': 'warn',
      '@typescript-eslint/no-unused-vars': ['warn', { argsIgnorePattern: '^_', varsIgnorePattern: '^_' }],
      '@typescript-eslint/no-explicit-any': 'off'
      // 'import/order': ['warn', { alphabetize: { order: 'asc', caseInsensitive: true }, 'newlines-between': 'always' }],
    },
  },
];
