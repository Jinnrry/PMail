# fe

前端代码库

```sh
yarn
```

### Compile and Hot-Reload for Development

```sh
yarn dev
```

### Compile and Minify for Production

```sh
yarn build
```

### Lint with [ESLint](https://eslint.org/)

```sh
yarn lint
```

### Browser Regression Tests

测试邮件正文的高度上限、沙箱脚本拦截和跨域链接导航。需要本机安装 Chrome 或 Chromium；测试工具不会自动下载浏览器。

```sh
yarn test
# 浏览器不在默认路径时：
CHROME_BIN=/path/to/chrome yarn test
```
