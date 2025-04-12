
**How to Build and Run:**

1.  **Save:** Save the code above as `packfiles.go`.
2.  **Tidy Dependencies:** Run `go mod tidy` in the directory where you saved the file. This creates/updates `go.mod` and `go.sum`.
3.  **Build:** Run `go build -o packfiles` (or `packfiles.exe` on Windows).
4.  **Run:** Execute the compiled program.

    *   **Inside a Git Repo:**
        ```bash
        ./packfiles > output.txt # Pack all tracked files
        ./packfiles -I "*.go" -I "*.mod" > go_files.txt # Pack only Go files
        ./packfiles -E "vendor/*" -E "*.log" > filtered_output.txt # Exclude vendor and logs
        ```
    *   **Outside a Git Repo or Specific Path:**
        ```bash
        ./packfiles --start-path /path/to/your/project > project_files.txt
        ./packfiles --start-path . > current_dir_files.txt # Pack current dir if not git
        ./packfiles --start-path /some/dir -E "temp/*" > specific_dir_filtered.txt
        ```

**Explanation:**

1.  **Flags:** Uses the standard `flag` package. A custom `stringSlice` type is defined to handle multiple occurrences of `--exclude`/`-E` and `--include`/`-I`.
2.  **Root Path Determination:**
    *   Checks if `--start-path` is provided. If yes, uses that path and forces "Filesystem Walk Mode". It validates that the path exists and is a directory.
    *   If no `--start-path`, it tries to find a `.git` directory upwards from the current working directory using `go-git`'s `PlainOpenWithOptions`.
    *   If a Git repo is found, it sets `useGit = true` and the `rootPath` to the repo's root.
    *   If no Git repo is found, it sets `useGit = false` and `rootPath` to the current working directory.
3.  **Git Mode (`useGit = true`):**
    *   Opens the Git repository.
    *   Gets the `HEAD` reference, the corresponding commit, and its file tree.
    *   Iterates through the files in the tree using `tree.Files()`.
    *   For each file (`object.File`):
        *   Gets the relative path (`f.Name`).
        *   Calculates the full path.
        *   **Crucially:** It checks `f.Mode.IsRegular()` to ensure it's a file, `!f.Mode.IsGitLink()` to skip submodules, and `!f.Mode.IsSymlink()` to skip symlinks tracked by Git. It also performs an `os.Lstat` to ensure the file actually exists in the working directory (it might be tracked but deleted locally).
        *   Calls `processFile` to handle filtering and output.
4.  **Filesystem Walk Mode (`useGit = false`):**
    *   Uses `filepath.WalkDir` (more modern than `filepath.Walk`) to traverse the directory tree starting from `rootPath`.
    *   The callback function receives the `path`, `DirEntry`, and `error`.
    *   Calculates the `relativePath`.
    *   Uses `d.Type()&fs.ModeSymlink != 0` to detect and skip symlinks.
    *   Uses `d.IsDir()` to skip directories (but allows traversal). An important optimization is added: if a directory *itself* matches an exclude pattern (like `vendor/` or `node_modules`), it returns `filepath.SkipDir` to prevent `WalkDir` from descending into it.
    *   Calls `processFile` for actual files.
5.  **`processFile` Function:**
    *   Takes the output writer, full path, and relative path.
    *   Performs a final `os.Lstat` check to be absolutely sure it's not processing a symlink.
    *   Calls `isIncluded` and `isExcluded` helper functions with the **relative path**.
    *   If the file passes filters:
        *   Writes a header comment like `--- // File: path/to/file.ext // ---`.
        *   Opens the file using the **full path**.
        *   Copies the file content to the output writer (`os.Stdout` in this case).
        *   Writes a trailing newline.
6.  **Filtering (`isExcluded`, `isIncluded`):**
    *   Use `filepath.Match` for glob pattern matching against the relative path.
    *   `isExcluded` checks all exclude patterns. Special handling is added to make directory patterns like `vendor/` or `testdata/*` exclude files *within* those directories correctly.
    *   `isIncluded` checks include patterns. If the `includePaths` slice is empty, it defaults to `true` (include everything). Otherwise, a file must match at least one include pattern.
7.  **Error Handling:** Uses `log.Fatalf` for critical errors that prevent the program from continuing and `log.Printf` for warnings (like invalid patterns or inability to read a specific file) allowing the program to skip problematic items and continue.
