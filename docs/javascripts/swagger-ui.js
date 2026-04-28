document$.subscribe(function() {
  if (document.getElementById("swagger-ui")) {
    const link = document.createElement("link");
    link.rel = "stylesheet";
    link.href = "https://cdn.jsdelivr.net/npm/swagger-ui-dist@4/swagger-ui.css";
    document.head.appendChild(link);

    const bundleScript = document.createElement("script");
    bundleScript.src = "https://cdn.jsdelivr.net/npm/swagger-ui-dist@4/swagger-ui-bundle.js";
    bundleScript.onload = function() {
      const standaloneScript = document.createElement("script");
      standaloneScript.src = "https://cdn.jsdelivr.net/npm/swagger-ui-dist@4/swagger-ui-standalone-preset.js";
      standaloneScript.onload = function() {
        SwaggerUIBundle({
          url: "../openapi.yml",
          dom_id: "#swagger-ui",
          presets: [
            SwaggerUIBundle.presets.apis,
            window.SwaggerUIStandalonePreset
          ],
          plugins: [
            SwaggerUIBundle.plugins.DownloadUrl
          ],
          layout: "StandaloneLayout",
          deepLinking: true
        });
      };
      document.head.appendChild(standaloneScript);
    };
    document.head.appendChild(bundleScript);
  }
});

