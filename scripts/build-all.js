#!/usr/bin/env node

const { spawn } = require('child_process');
const fs = require('fs');
const path = require('path');

// Only build for current platform during development
// Full cross-platform builds should be done in CI/CD
const currentPlatform = {
  os: process.platform === 'win32' ? 'windows' : process.platform,
  arch: process.arch === 'x64' ? 'amd64' : process.arch,
  ext: process.platform === 'win32' ? '.exe' : ''
};

const platforms = process.env.BUILD_ALL_PLATFORMS === 'true'
  ? [
      { os: 'darwin', arch: 'amd64', ext: '' },
      { os: 'darwin', arch: 'arm64', ext: '' },
      { os: 'linux', arch: 'amd64', ext: '' },
      { os: 'linux', arch: 'arm64', ext: '' },
      { os: 'windows', arch: 'amd64', ext: '.exe' },
      { os: 'windows', arch: 'arm64', ext: '.exe' }
    ]
  : [currentPlatform];

async function buildBinary(platform) {
  const outputName = `nestjs-module-lint-${platform.os}-${platform.arch}${platform.ext}`;
  const outputPath = path.join(__dirname, '..', 'dist', outputName);

  console.log(`Building for ${platform.os}/${platform.arch}...`);

  return new Promise((resolve, reject) => {
    const env = {
      ...process.env,
      GOOS: platform.os,
      GOARCH: platform.arch
    };

    const build = spawn('go', ['build', '-o', outputPath, 'main.go'], {
      env,
      cwd: path.join(__dirname, '..'),
      stdio: 'inherit'
    });

    build.on('close', (code) => {
      if (code !== 0) {
        reject(new Error(`Build failed for ${platform.os}/${platform.arch} with code ${code}`));
      } else {
        console.log(`✓ Built ${outputName}`);
        resolve();
      }
    });

    build.on('error', (err) => {
      reject(err);
    });
  });
}

async function buildAll() {
  // Create dist directory
  const distDir = path.join(__dirname, '..', 'dist');
  if (!fs.existsSync(distDir)) {
    fs.mkdirSync(distDir);
  }

  console.log('Building binaries for all platforms...\n');

  for (const platform of platforms) {
    try {
      await buildBinary(platform);
    } catch (error) {
      console.error(`Failed to build for ${platform.os}/${platform.arch}:`, error.message);
      // Continue building other platforms
    }
  }

  console.log('\nBuild complete!');
}

// Only run if this is being executed directly
if (require.main === module) {
  buildAll().catch((error) => {
    console.error('Build failed:', error);
    process.exit(1);
  });
}