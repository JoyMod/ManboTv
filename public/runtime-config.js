// 运行时配置
// 此文件在构建后可以被修改，用于配置 API 地址等
(function() {
  // 从环境变量读取 (如果在 Docker 中，可以通过注入此文件来修改配置)
  const apiBaseUrl = '${API_BASE_URL}' || '';
  
  window.RUNTIME_CONFIG = {
    // API 基础 URL
    // 留空表示使用同域 (Nginx 反向代理到 Go 后端)
    // 设置为完整 URL 如: 'http://localhost:8080' 可直接访问后端
    API_BASE_URL: apiBaseUrl.replace('${API_BASE_URL}', ''),
    
    // 站点名称
    SITE_NAME: 'ManboTV',
    
    // 豆瓣代理类型
    DOUBAN_PROXY_TYPE: 'cmliussss-cdn-tencent',
    DOUBAN_IMAGE_PROXY_TYPE: 'cmliussss-cdn-tencent',
    
    // 存储类型
    STORAGE_TYPE: 'redis',
  };
})();
