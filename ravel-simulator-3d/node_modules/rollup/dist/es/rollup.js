/*
  @license
	Rollup.js v4.53.3
	Wed, 19 Nov 2025 06:31:27 GMT - commit 998b5950a6ea7cea1a7b994e8dab45472c3cbe7e

	https://github.com/rollup/rollup

	Released under the MIT License.
*/
export { version as VERSION, defineConfig, rollup, watch } from './shared/node-entry.js';
import './shared/parseAst.js';
import '../native.js';
import 'node:path';
import 'path';
import 'node:process';
import 'node:perf_hooks';
import 'node:fs/promises';
