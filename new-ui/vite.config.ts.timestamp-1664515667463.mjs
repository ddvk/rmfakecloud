var __create = Object.create;
var __defProp = Object.defineProperty;
var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
var __getOwnPropNames = Object.getOwnPropertyNames;
var __getProtoOf = Object.getPrototypeOf;
var __hasOwnProp = Object.prototype.hasOwnProperty;
var __commonJS = (cb, mod) => function __require() {
  return mod || (0, cb[__getOwnPropNames(cb)[0]])((mod = { exports: {} }).exports, mod), mod.exports;
};
var __copyProps = (to, from, except, desc) => {
  if (from && typeof from === "object" || typeof from === "function") {
    for (let key of __getOwnPropNames(from))
      if (!__hasOwnProp.call(to, key) && key !== except)
        __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
  }
  return to;
};
var __toESM = (mod, isNodeMode, target) => (target = mod != null ? __create(__getProtoOf(mod)) : {}, __copyProps(
  isNodeMode || !mod || !mod.__esModule ? __defProp(target, "default", { value: mod, enumerable: true }) : target,
  mod
));

// tailwind.config.cjs
var require_tailwind_config = __commonJS({
  "tailwind.config.cjs"(exports, module) {
    "use strict";
    module.exports = {
      content: ["./*.html", "./src/**/*.{js,ts,jsx,tsx,css}"],
      theme: {
        extend: {
          keyframes: {
            "roll-down": {
              "0%": { "max-height": "0" },
              "100%": { "max-height": "100vh" }
            },
            fadein: {
              from: { opacity: 0 },
              to: { opacity: 1 }
            },
            "roll-up": {
              "0%": { opacity: 1, "max-height": "100vh" },
              "100%": { opacity: 0, "max-height": "0" }
            },
            "flip-x": {
              "0%": { transform: "rotateX(180deg)" },
              "100%": { transform: "rotateX(0deg)" }
            },
            "flip-x-reverse": {
              "0%": { transform: "rotateX(-180deg)" },
              "100%": { transform: "rotateX(0deg)" }
            },
            slidein: {
              "0%": { transform: "translateX(100%)", opacity: 0 },
              "100%": { transform: "translateX(0)", opacity: 1 }
            },
            slideout: {
              "0%": { transform: "translateX(0)", opacity: 1 },
              "100%": { transform: "translateX(100%)", opacity: 0 }
            }
          },
          animation: {
            "roll-down": "roll-down 0.5s ease-in",
            "roll-up": "roll-up 0.5s ease-out",
            fadein: "fadein 0.5s ease-in",
            "flip-x": "flip-x 0.3s ease-out",
            "flip-x-reverse": "flip-x-reverse 0.3s ease-out"
          }
        },
        fontFamily: {
          sans: ["system-ui"]
        }
      },
      plugins: []
    };
  }
});

// vite.config.ts
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import { VitePluginFonts } from "vite-plugin-fonts";
import { createHtmlPlugin } from "vite-plugin-html";
import viteImagemin from "vite-plugin-imagemin";
import { viteStaticCopy } from "vite-plugin-static-copy";

// config.js
var config = {
  imagemin: {
    gifsicle: {
      optimizationLevel: 7,
      interlaced: false
    },
    webp: {
      quality: 75
    },
    optipng: {
      optimizationLevel: 7
    },
    mozjpeg: {
      quality: 20
    },
    pngquant: {
      quality: [0.8, 0.9],
      speed: 4
    },
    svgo: {
      plugins: [
        {
          name: "removeViewBox"
        },
        {
          name: "removeStyleElement",
          active: true
        }
      ]
    }
  }
};
var config_default = config;

// postcss.config.js
var import_tailwind_config = __toESM(require_tailwind_config(), 1);
import autoprefixer from "autoprefixer";
import tailwind from "tailwindcss";
var postcss_config_default = {
  plugins: [tailwind(import_tailwind_config.default), autoprefixer]
};

// vite.config.ts
var { imagemin } = config_default;
var vite_config_default = defineConfig({
  server: {
    proxy: {
      "/ui/api": {
        target: "http://localhost:3000",
        changeOrigin: true
      }
    }
  },
  plugins: [
    react(),
    viteImagemin(imagemin),
    viteStaticCopy({
      targets: [
        {
          src: "node_modules/pdfjs-dist/build/pdf.worker.min.js",
          dest: "lib/"
        },
        {
          src: "node_modules/pdfjs-dist/cmaps/",
          dest: "lib/"
        }
      ]
    }),
    createHtmlPlugin({
      minify: true,
      entry: "/src/main.tsx"
    }),
    VitePluginFonts({
      custom: {
        families: [
          {
            name: "CascadiaCodePL",
            src: "./src/assets/fonts/*.woff2"
          }
        ],
        display: "swap",
        preload: true,
        prefetch: false,
        injectTo: "head-prepend"
      }
    })
  ],
  css: {
    postcss: postcss_config_default
  },
  resolve: {
    alias: [
      { find: "@/", replacement: "/src" },
      { find: "@/Assets", replacement: "/src/assets" },
      { find: "@/Components", replacement: "/src/components" },
      { find: "@/API", replacement: "/src/api" }
    ]
  }
});
export {
  vite_config_default as default
};
//# sourceMappingURL=data:application/json;base64,ewogICJ2ZXJzaW9uIjogMywKICAic291cmNlcyI6IFsidGFpbHdpbmQuY29uZmlnLmNqcyIsICJ2aXRlLmNvbmZpZy50cyIsICJjb25maWcuanMiLCAicG9zdGNzcy5jb25maWcuanMiXSwKICAic291cmNlc0NvbnRlbnQiOiBbImNvbnN0IF9fdml0ZV9pbmplY3RlZF9vcmlnaW5hbF9kaXJuYW1lID0gXCIvaG9tZS9kZXZib29rL2RldmVsb3BtZW50L3dlYi9ybWZha2VjbG91ZC9uZXctdWlcIjtjb25zdCBfX3ZpdGVfaW5qZWN0ZWRfb3JpZ2luYWxfZmlsZW5hbWUgPSBcIi9ob21lL2RldmJvb2svZGV2ZWxvcG1lbnQvd2ViL3JtZmFrZWNsb3VkL25ldy11aS90YWlsd2luZC5jb25maWcuY2pzXCI7Y29uc3QgX192aXRlX2luamVjdGVkX29yaWdpbmFsX2ltcG9ydF9tZXRhX3VybCA9IFwiZmlsZTovLy9ob21lL2RldmJvb2svZGV2ZWxvcG1lbnQvd2ViL3JtZmFrZWNsb3VkL25ldy11aS90YWlsd2luZC5jb25maWcuY2pzXCI7LyoqIEB0eXBlIHtpbXBvcnQoJ3RhaWx3aW5kY3NzJykuQ29uZmlnfSAqL1xuXG5tb2R1bGUuZXhwb3J0cyA9IHtcbiAgY29udGVudDogWycuLyouaHRtbCcsICcuL3NyYy8qKi8qLntqcyx0cyxqc3gsdHN4LGNzc30nXSxcbiAgdGhlbWU6IHtcbiAgICBleHRlbmQ6IHtcbiAgICAgIGtleWZyYW1lczoge1xuICAgICAgICAncm9sbC1kb3duJzoge1xuICAgICAgICAgICcwJSc6IHsgJ21heC1oZWlnaHQnOiAnMCcgfSxcbiAgICAgICAgICAnMTAwJSc6IHsgJ21heC1oZWlnaHQnOiAnMTAwdmgnIH1cbiAgICAgICAgfSxcbiAgICAgICAgZmFkZWluOiB7XG4gICAgICAgICAgZnJvbTogeyBvcGFjaXR5OiAwIH0sXG4gICAgICAgICAgdG86IHsgb3BhY2l0eTogMSB9XG4gICAgICAgIH0sXG4gICAgICAgICdyb2xsLXVwJzoge1xuICAgICAgICAgICcwJSc6IHsgb3BhY2l0eTogMSwgJ21heC1oZWlnaHQnOiAnMTAwdmgnIH0sXG4gICAgICAgICAgJzEwMCUnOiB7IG9wYWNpdHk6IDAsICdtYXgtaGVpZ2h0JzogJzAnIH1cbiAgICAgICAgfSxcbiAgICAgICAgJ2ZsaXAteCc6IHtcbiAgICAgICAgICAnMCUnOiB7IHRyYW5zZm9ybTogJ3JvdGF0ZVgoMTgwZGVnKScgfSxcbiAgICAgICAgICAnMTAwJSc6IHsgdHJhbnNmb3JtOiAncm90YXRlWCgwZGVnKScgfVxuICAgICAgICB9LFxuICAgICAgICAnZmxpcC14LXJldmVyc2UnOiB7XG4gICAgICAgICAgJzAlJzogeyB0cmFuc2Zvcm06ICdyb3RhdGVYKC0xODBkZWcpJyB9LFxuICAgICAgICAgICcxMDAlJzogeyB0cmFuc2Zvcm06ICdyb3RhdGVYKDBkZWcpJyB9XG4gICAgICAgIH0sXG4gICAgICAgIHNsaWRlaW46IHtcbiAgICAgICAgICAnMCUnOiB7IHRyYW5zZm9ybTogJ3RyYW5zbGF0ZVgoMTAwJSknLCBvcGFjaXR5OiAwIH0sXG4gICAgICAgICAgJzEwMCUnOiB7IHRyYW5zZm9ybTogJ3RyYW5zbGF0ZVgoMCknLCBvcGFjaXR5OiAxIH1cbiAgICAgICAgfSxcbiAgICAgICAgc2xpZGVvdXQ6IHtcbiAgICAgICAgICAnMCUnOiB7IHRyYW5zZm9ybTogJ3RyYW5zbGF0ZVgoMCknLCBvcGFjaXR5OiAxIH0sXG4gICAgICAgICAgJzEwMCUnOiB7IHRyYW5zZm9ybTogJ3RyYW5zbGF0ZVgoMTAwJSknLCBvcGFjaXR5OiAwIH1cbiAgICAgICAgfVxuICAgICAgfSxcbiAgICAgIGFuaW1hdGlvbjoge1xuICAgICAgICAncm9sbC1kb3duJzogJ3JvbGwtZG93biAwLjVzIGVhc2UtaW4nLFxuICAgICAgICAncm9sbC11cCc6ICdyb2xsLXVwIDAuNXMgZWFzZS1vdXQnLFxuICAgICAgICBmYWRlaW46ICdmYWRlaW4gMC41cyBlYXNlLWluJyxcbiAgICAgICAgJ2ZsaXAteCc6ICdmbGlwLXggMC4zcyBlYXNlLW91dCcsXG4gICAgICAgICdmbGlwLXgtcmV2ZXJzZSc6ICdmbGlwLXgtcmV2ZXJzZSAwLjNzIGVhc2Utb3V0J1xuICAgICAgfVxuICAgIH0sXG4gICAgZm9udEZhbWlseToge1xuICAgICAgc2FuczogWydzeXN0ZW0tdWknXVxuICAgIH1cbiAgfSxcbiAgcGx1Z2luczogW11cbn1cbiIsICJjb25zdCBfX3ZpdGVfaW5qZWN0ZWRfb3JpZ2luYWxfZGlybmFtZSA9IFwiL2hvbWUvZGV2Ym9vay9kZXZlbG9wbWVudC93ZWIvcm1mYWtlY2xvdWQvbmV3LXVpXCI7Y29uc3QgX192aXRlX2luamVjdGVkX29yaWdpbmFsX2ZpbGVuYW1lID0gXCIvaG9tZS9kZXZib29rL2RldmVsb3BtZW50L3dlYi9ybWZha2VjbG91ZC9uZXctdWkvdml0ZS5jb25maWcudHNcIjtjb25zdCBfX3ZpdGVfaW5qZWN0ZWRfb3JpZ2luYWxfaW1wb3J0X21ldGFfdXJsID0gXCJmaWxlOi8vL2hvbWUvZGV2Ym9vay9kZXZlbG9wbWVudC93ZWIvcm1mYWtlY2xvdWQvbmV3LXVpL3ZpdGUuY29uZmlnLnRzXCI7LyoqIEB0eXBlIHtpbXBvcnQoJ3ZpdGUnKS5Vc2VyQ29uZmlnfSAqL1xuXG5pbXBvcnQgcmVhY3QgZnJvbSAnQHZpdGVqcy9wbHVnaW4tcmVhY3QnXG5pbXBvcnQgeyBkZWZpbmVDb25maWcgfSBmcm9tICd2aXRlJ1xuaW1wb3J0IHsgVml0ZVBsdWdpbkZvbnRzIH0gZnJvbSAndml0ZS1wbHVnaW4tZm9udHMnXG5pbXBvcnQgeyBjcmVhdGVIdG1sUGx1Z2luIH0gZnJvbSAndml0ZS1wbHVnaW4taHRtbCdcbmltcG9ydCB2aXRlSW1hZ2VtaW4gZnJvbSAndml0ZS1wbHVnaW4taW1hZ2VtaW4nXG5pbXBvcnQgeyB2aXRlU3RhdGljQ29weSB9IGZyb20gJ3ZpdGUtcGx1Z2luLXN0YXRpYy1jb3B5J1xuXG5pbXBvcnQgY29uZmlnIGZyb20gJy4vY29uZmlnLmpzJ1xuaW1wb3J0IHBvc3Rjc3MgZnJvbSAnLi9wb3N0Y3NzLmNvbmZpZy5qcydcblxuY29uc3QgeyBpbWFnZW1pbiB9ID0gY29uZmlnXG5cbi8vIGh0dHBzOi8vdml0ZWpzLmRldi9jb25maWcvXG5leHBvcnQgZGVmYXVsdCBkZWZpbmVDb25maWcoe1xuICBzZXJ2ZXI6IHtcbiAgICBwcm94eToge1xuICAgICAgJy91aS9hcGknOiB7XG4gICAgICAgIHRhcmdldDogJ2h0dHA6Ly9sb2NhbGhvc3Q6MzAwMCcsXG4gICAgICAgIGNoYW5nZU9yaWdpbjogdHJ1ZVxuICAgICAgfVxuICAgIH1cbiAgfSxcbiAgcGx1Z2luczogW1xuICAgIHJlYWN0KCksXG4gICAgdml0ZUltYWdlbWluKGltYWdlbWluKSxcbiAgICB2aXRlU3RhdGljQ29weSh7XG4gICAgICB0YXJnZXRzOiBbXG4gICAgICAgIHtcbiAgICAgICAgICBzcmM6ICdub2RlX21vZHVsZXMvcGRmanMtZGlzdC9idWlsZC9wZGYud29ya2VyLm1pbi5qcycsXG4gICAgICAgICAgZGVzdDogJ2xpYi8nXG4gICAgICAgIH0sXG4gICAgICAgIHtcbiAgICAgICAgICBzcmM6ICdub2RlX21vZHVsZXMvcGRmanMtZGlzdC9jbWFwcy8nLFxuICAgICAgICAgIGRlc3Q6ICdsaWIvJ1xuICAgICAgICB9XG4gICAgICBdXG4gICAgfSksXG4gICAgY3JlYXRlSHRtbFBsdWdpbih7XG4gICAgICBtaW5pZnk6IHRydWUsXG4gICAgICBlbnRyeTogJy9zcmMvbWFpbi50c3gnXG4gICAgfSksXG4gICAgVml0ZVBsdWdpbkZvbnRzKHtcbiAgICAgIC8vIEN1c3RvbSBmb250c1xuICAgICAgY3VzdG9tOiB7XG4gICAgICAgIGZhbWlsaWVzOiBbXG4gICAgICAgICAge1xuICAgICAgICAgICAgbmFtZTogJ0Nhc2NhZGlhQ29kZVBMJyxcbiAgICAgICAgICAgIHNyYzogJy4vc3JjL2Fzc2V0cy9mb250cy8qLndvZmYyJ1xuICAgICAgICAgIH1cbiAgICAgICAgXSxcbiAgICAgICAgZGlzcGxheTogJ3N3YXAnLFxuICAgICAgICBwcmVsb2FkOiB0cnVlLFxuICAgICAgICBwcmVmZXRjaDogZmFsc2UsXG4gICAgICAgIGluamVjdFRvOiAnaGVhZC1wcmVwZW5kJ1xuICAgICAgfVxuICAgIH0pXG4gIF0sXG4gIGNzczoge1xuICAgIHBvc3Rjc3NcbiAgfSxcbiAgcmVzb2x2ZToge1xuICAgIGFsaWFzOiBbXG4gICAgICB7IGZpbmQ6ICdALycsIHJlcGxhY2VtZW50OiAnL3NyYycgfSxcbiAgICAgIHsgZmluZDogJ0AvQXNzZXRzJywgcmVwbGFjZW1lbnQ6ICcvc3JjL2Fzc2V0cycgfSxcbiAgICAgIHsgZmluZDogJ0AvQ29tcG9uZW50cycsIHJlcGxhY2VtZW50OiAnL3NyYy9jb21wb25lbnRzJyB9LFxuICAgICAgeyBmaW5kOiAnQC9BUEknLCByZXBsYWNlbWVudDogJy9zcmMvYXBpJyB9XG4gICAgXVxuICB9XG59KVxuIiwgImNvbnN0IF9fdml0ZV9pbmplY3RlZF9vcmlnaW5hbF9kaXJuYW1lID0gXCIvaG9tZS9kZXZib29rL2RldmVsb3BtZW50L3dlYi9ybWZha2VjbG91ZC9uZXctdWlcIjtjb25zdCBfX3ZpdGVfaW5qZWN0ZWRfb3JpZ2luYWxfZmlsZW5hbWUgPSBcIi9ob21lL2RldmJvb2svZGV2ZWxvcG1lbnQvd2ViL3JtZmFrZWNsb3VkL25ldy11aS9jb25maWcuanNcIjtjb25zdCBfX3ZpdGVfaW5qZWN0ZWRfb3JpZ2luYWxfaW1wb3J0X21ldGFfdXJsID0gXCJmaWxlOi8vL2hvbWUvZGV2Ym9vay9kZXZlbG9wbWVudC93ZWIvcm1mYWtlY2xvdWQvbmV3LXVpL2NvbmZpZy5qc1wiO2NvbnN0IGNvbmZpZyA9IHtcbiAgaW1hZ2VtaW46IHtcbiAgICBnaWZzaWNsZToge1xuICAgICAgb3B0aW1pemF0aW9uTGV2ZWw6IDcsXG4gICAgICBpbnRlcmxhY2VkOiBmYWxzZVxuICAgIH0sXG4gICAgd2VicDoge1xuICAgICAgcXVhbGl0eTogNzVcbiAgICB9LFxuICAgIG9wdGlwbmc6IHtcbiAgICAgIG9wdGltaXphdGlvbkxldmVsOiA3XG4gICAgfSxcbiAgICBtb3pqcGVnOiB7XG4gICAgICBxdWFsaXR5OiAyMFxuICAgIH0sXG4gICAgcG5ncXVhbnQ6IHtcbiAgICAgIHF1YWxpdHk6IFswLjgsIDAuOV0sXG4gICAgICBzcGVlZDogNFxuICAgIH0sXG4gICAgc3Znbzoge1xuICAgICAgcGx1Z2luczogW1xuICAgICAgICB7XG4gICAgICAgICAgbmFtZTogJ3JlbW92ZVZpZXdCb3gnXG4gICAgICAgIH0sXG4gICAgICAgIHtcbiAgICAgICAgICBuYW1lOiAncmVtb3ZlU3R5bGVFbGVtZW50JyxcbiAgICAgICAgICBhY3RpdmU6IHRydWVcbiAgICAgICAgfVxuICAgICAgXVxuICAgIH1cbiAgfVxufVxuXG5leHBvcnQgZGVmYXVsdCBjb25maWdcbiIsICJjb25zdCBfX3ZpdGVfaW5qZWN0ZWRfb3JpZ2luYWxfZGlybmFtZSA9IFwiL2hvbWUvZGV2Ym9vay9kZXZlbG9wbWVudC93ZWIvcm1mYWtlY2xvdWQvbmV3LXVpXCI7Y29uc3QgX192aXRlX2luamVjdGVkX29yaWdpbmFsX2ZpbGVuYW1lID0gXCIvaG9tZS9kZXZib29rL2RldmVsb3BtZW50L3dlYi9ybWZha2VjbG91ZC9uZXctdWkvcG9zdGNzcy5jb25maWcuanNcIjtjb25zdCBfX3ZpdGVfaW5qZWN0ZWRfb3JpZ2luYWxfaW1wb3J0X21ldGFfdXJsID0gXCJmaWxlOi8vL2hvbWUvZGV2Ym9vay9kZXZlbG9wbWVudC93ZWIvcm1mYWtlY2xvdWQvbmV3LXVpL3Bvc3Rjc3MuY29uZmlnLmpzXCI7aW1wb3J0IGF1dG9wcmVmaXhlciBmcm9tICdhdXRvcHJlZml4ZXInXG5pbXBvcnQgdGFpbHdpbmQgZnJvbSAndGFpbHdpbmRjc3MnXG5cbmltcG9ydCB0YWlsd2luZENvbmZpZyBmcm9tICcuL3RhaWx3aW5kLmNvbmZpZy5janMnXG5cbmV4cG9ydCBkZWZhdWx0IHtcbiAgcGx1Z2luczogW3RhaWx3aW5kKHRhaWx3aW5kQ29uZmlnKSwgYXV0b3ByZWZpeGVyXVxufVxuIl0sCiAgIm1hcHBpbmdzIjogIjs7Ozs7Ozs7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
