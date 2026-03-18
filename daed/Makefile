name: Build and Release daed

on:
  schedule:
    - cron: '*/30 * * * *'
  workflow_dispatch:

concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        include:
          - goarch: amd64
            arch_name: x86_64
            gcc_pkg: gcc-multilib
          - goarch: arm64
            arch_name: aarch64
            gcc_pkg: gcc-aarch64-linux-gnu
    outputs:
      short_sha: ${{ steps.vars.outputs.short_sha }}
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          submodules: recursive
          fetch-depth: 0
          token: ${{ secrets.ACTIONS_DEPLOY_KEY }}

      - name: Set Short SHA
        id: vars
        run: echo "short_sha=$(git rev-parse --short HEAD)" >> $GITHUB_OUTPUT

      - name: Setup Node
        uses: actions/setup-node@v4
        with:
          node-version: 20

      - name: Setup PNPM
        uses: pnpm/action-setup@v3
        with:
          version: 9
          run_install: false

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true

      - name: Install System Deps
        run: |
          sudo apt-get update
          sudo apt-get install -y build-essential libpcap-dev ${{ matrix.gcc_pkg }}

      - name: Patch dae-core
        run: |
          # 根据 Makefile，确认后端目录为 wing
          if [ -d "wing" ]; then
            cd wing
            git remote add patch_repo https://github.com/olicesx/dae.git || true
            git fetch patch_repo optimize/code-quality-fixes
            git checkout patch_repo/optimize/code-quality-fixes
            cd ..
          else
            echo "Warning: wing directory not found, skipping patch"
          fi

      - name: Build Binary
        env:
          GOOS: linux
          GOARCH: ${{ matrix.goarch }}
          CGO_ENABLED: 1
          CC: ${{ matrix.goarch == 'arm64' && 'aarch64-linux-gnu-gcc' || 'gcc' }}
          SKIP_SUBMODULES: 1
        run: |
          # 直接调用 make daed，它会自动触发 dist 和后端编译
          make daed VERSION=${{ steps.vars.outputs.short_sha }} OUTPUT=daed-${{ matrix.arch_name }}

      - name: Upload Artifact
        uses: actions/upload-artifact@v4
        with:
          name: daed-${{ matrix.arch_name }}
          path: daed-${{ matrix.arch_name }}

  release:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - name: Download Artifacts
        uses: actions/download-artifact@v4

      - name: Create Release
        uses: softprops/action-gh-release@v2
        with:
          token: ${{ secrets.ACTIONS_DEPLOY_KEY }}
          tag_name: kix-daed_${{ needs.build.outputs.short_sha }}
          name: Release kix-daed_${{ needs.build.outputs.short_sha }}
          files: |
            daed-x86_64/daed-x86_64
            daed-aarch64/daed-aarch64

  cleanup:
    needs: release
    if: always()
    runs-on: ubuntu-latest
    steps:
      - name: Cleanup All
        env:
          GH_TOKEN: ${{ secrets.ACTIONS_DEPLOY_KEY }}
        run: |
          # 清理旧 Release
          gh release list --limit 100 | grep "kix-daed_" | tail -n +4 | awk '{print $1}' | x86 -I {} gh release delete {} --yes --cleanup-tag || true
          
          # 清理 3 天前的 Workflow Runs
          CUTOFF=$(date -d '3 days ago' +%Y-%m-%d)
          for wf in $(gh workflow list --json id -q '.[].id'); do
            gh run delete --workflow "$wf" --created "<$CUTOFF" --confirm || true
          done
