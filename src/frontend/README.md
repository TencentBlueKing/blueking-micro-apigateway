## 开发环境

- 使用 Node.js 24.x，最低版本为 24.22.1；Docker 构建和 `.nvmrc` 固定使用 24.22.1。
- 此版本包含 CVE-2026-111497 的修复，详见 [Node.js 安全公告](https://nodejs.org/en/blog/vulnerability/october-2026-security-releases)。
- 使用 nvm 时，先运行 `nvm install` 和 `nvm use`；nvm-windows 用户运行 `nvm install 24.22.1` 和 `nvm use 24.22.1`。

```shell
node --version
npm install --legacy-peer-deps
npm run build:ee
npm run lint -- --no-fix
npm run test:schema
```

现有 UI 组件存在 peer dependency 版本冲突，安装时使用 `--legacy-peer-deps`，与 Docker 构建保持一致。

## 微网关前端开发注意事项

### 1. `lodash` 改用 `lodash-es`

引入时请使用 `import _ from 'lodash-es'`

### 2. 每个模块的路由单据放到一个 ts 文件里，再导入到 `router/index.ts` 中，不再放到一个文件里

```typescript
// router/index.ts
const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: Home,
    redirect: routeRoutes[0].path,
    children: [
      ...routeRoutes,
      ...serviceRoutes,
      ...upstreamRoutes,
    ],
  },
];
```

### 3. 每个页面的首页不用 `xxx/index.vue` 命名

如：

以前是：`home/index.vue`，`resource/index.vue`

改为： `home/home.vue`，`resource/resource.vue`

避免过多的 `index.vue`，编辑器的 tab 开多了难以辨认，同时可以通过 url 直接检索到对应页面文件

### 4. commit 前通过 pre-commit hook 执行一个 eslint 检查

在项目根目录 `.git/hooks` 中创建一个 `pre-commit` 文件，添加以下内容，可以在每次提交前执行 eslint 检查，如果检查不通过则不允许提交

```shell
#!/bin/sh

STAGED_FILES=$(git diff --cached --name-only HEAD | grep -E '\.(js|jsx|ts|tsx|vue)$' | xargs)

echo "$STAGED_FILES"

if [[ "$STAGED_FILES" = "" ]]; then
  exit 0
fi

PASS=true

ESLINT="./src/frontend/node_modules/.bin/eslint"

echo "Running 'eslint [file]' on committed .js|.jsx|.ts|.tsx|.vue files in src/frontend"

if [[ ! -x "$ESLINT" ]]; then
  echo "Please install ESlint"
  exit 1
fi

for FILE in $STAGED_FILES
do
  "$ESLINT" --color "$FILE"

  if [[ "$?" == 0 ]]; then
    echo "ESLint Passed: $FILE"
  else
    echo "ESLint Failed: $FILE"
    PASS=false
  fi
done

echo "Lint finished"

if ! $PASS; then
  echo "COMMIT FAILED: Your commit contains files that failed the linting. Please run 'eslint --fix src/frontend' and try again."
  exit 1
else
  echo "COMMIT SUCCEEDED"
fi

exit $?
```
