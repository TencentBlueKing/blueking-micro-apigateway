const mockServer = require('./mock-server');
const MonacoWebpackPlugin = require('monaco-editor-webpack-plugin');

module.exports = {
  host: process.env.BK_APP_HOST,
  port: process.env.BK_APP_PORT,
  publicPath: '/',
  cache: true,
  open: true,
  replaceStatic: true,
  server: 'https',

  // webpack config 配置
  configureWebpack() {
    return {
      devServer: {
        setupMiddlewares: mockServer,
        host: 'dev-t.paas3-dev.bktencent.com',
        client: {
          overlay: false,
        },
        historyApiFallback: { rewrites: [{ from: /(.*?)\//, to: '/index.html' }] },
        // https: !process.env.BK_HTTPS,
      },
      // plugins: [
      //   require('unplugin-auto-import/webpack')
      //     .default({
      //       // targets to transform
      //       include: [
      //         /\.[tj]sx?$/, // .ts, .tsx, .js, .jsx
      //         /\.vue$/,
      //         /\.vue\?vue/, // .vue
      //       ],
      //
      //       // global imports to register
      //       imports: [
      //         // presets
      //         'vue',
      //         'vue-router',
      //         '@vueuse/core',
      //         {
      //           'vue-i18n': ['useI18n', 'createI18n'],
      //         },
      //       ],
      //       dts: true,
      //     }),
      // ],
    };
  },

  chainWebpack: (config) => {
    config
      .plugin('monaco')
      .use(MonacoWebpackPlugin);
    return config;
  },
};
