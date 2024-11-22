import { defineConfig} from 'vite'
import react from '@vitejs/plugin-react-swc'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  build: {
    rollupOptions: {
      // external:["pdfjs-dist"],
      output: {
        manualChunks: function (id) {
          // console.log(id);
          if (id.includes("node_modules")) {
            if (id.includes("pdf")) return "pdf";
            if (id.includes("react")) return "react";
            return "vendor";
          }
        },
      },
    },
  },
  css: {
    preprocessorOptions: {
      scss: {
        api: "modern-compiler",
        silenceDeprecations: [
          "mixed-decls",
          "color-functions",
          "global-builtin",
          "import",
        ],
      },
    },
  },
  server: {
	open: true,
	port: 3001,
    proxy: {
      "/ui/api": {
		target: "http://localhost:3000",
		changeOrigin: true
	  }
    },
  },
});
