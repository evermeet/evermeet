module.exports = [{
    script: './index.js',
    cwd: './packages/api',
    name: 'api',
  }, {
    script: './build/index.js',
    cwd: './packages/web',
    env: {
      PORT: 3001
    },
    name: 'web'
  }]