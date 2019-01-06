import svelte from 'rollup-plugin-svelte'
import resolve from 'rollup-plugin-node-resolve'
import commonjs from 'rollup-plugin-commonjs'
import nodent from 'rollup-plugin-nodent'
import buble from 'rollup-plugin-buble'
import uglify from 'rollup-plugin-uglify'
import scss from 'rollup-plugin-scss'

const production = !process.env.ROLLUP_WATCH

export default {
  input: 'src/main.js',
  output: {
    sourcemap: false,
    format: 'iife',
    name: 'app',
    file: 'public/bundle.js'
  },
  plugins: [
    svelte({
      dev: !production,
      store: true,
      cascade: false
    }),
    resolve(),
    commonjs(),
    scss({ output: 'public/bundle.css' }),
    production && nodent({
      promises: true,
      noRuntime: true
    }),
    production && buble({
      objectAssign: 'Object.assign',
      target: {chrome: 48, firefox: 44},
      transforms: {arrow: true, modules: false, dangerousForOf: false, generator: false}
    }),
    production && uglify()
  ]
}
