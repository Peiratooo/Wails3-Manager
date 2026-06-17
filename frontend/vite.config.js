import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'

function stripDaybrushPureAnnotations() {
  return {
    name: 'strip-daybrush-pure-annotations',
    enforce: 'pre',
    transform(code, id) {
      if (!id.includes('@daybrush/utils') || !id.endsWith('utils.esm.js')) {
        return null
      }
      return {
        code: code.replaceAll('/*#__PURE__*/', ''),
        map: null,
      }
    },
  }
}

// https://vite.dev/config/
export default defineConfig({
  plugins: [stripDaybrushPureAnnotations(), vue()],
})
