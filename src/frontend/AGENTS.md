# AGENTS.md

## Code Architecture

**Technology Stack**: Vue 3 + TypeScript + Vue Router + Pinia + BKUI-Vue

```plaintext
src/
├── views/          # Page components (20+ modules)
│   ├── route/      # Route management pages
│   ├── service/    # Service management pages
│   ├── upstream/   # Upstream management pages
│   └── ...         # Other resource management pages
├── router/         # Route definitions (split by module)
│   ├── index.ts    # Main router configuration
│   ├── route.ts    # Route module routes
│   ├── service.ts  # Service module routes
│   └── ...
├── store/          # Pinia stores
├── components/     # Reusable components
├── http/           # API client and request handling
├── hooks/          # Vue composables
└── i18n/           # Internationalization
```

**Key Conventions**:

1. **No index.vue**: Page components use descriptive names (e.g., `route/route.vue`, not `route/index.vue`) for better tab identification
2. **Modular Routing**: Each module's routes are in separate files, then imported into `router/index.ts`
3. **Use lodash-es**: Import as `import _ from 'lodash-es'` (not `lodash`)
4. **Linting**: Run ESLint (for example, via `npm run lint`) before committing frontend changes
