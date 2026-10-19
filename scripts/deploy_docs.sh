#!/usr/bin/env bash
set -euo pipefail

# deploy_docs.sh publishes the contents of the /docs directory
# to the gh-pages branch, which GitHub Pages serves from. Docs
# built from the main branch are placed at the root of the site
# while docs built from any other branch are placed under a
# subdirectory named after the branch.
#
# Usage:
#   deploy_docs.sh <branch>           Publish /docs for <branch>.
#   deploy_docs.sh --remove <branch>  Remove the docs for <branch>.
#
# The /docs directory should already contain the generated HTML
# (see 'build.sh --docs'). The 'origin' remote must be pushable.

# Change directory to repo root.
cd "${0%/*}/.."

MAIN_BRANCH="main"
PAGES_BRANCH="gh-pages"
DOMAIN="jander.land"
ATTEMPTS=5

# Top-level entries of the site owned by the main branch. A
# branch whose subdirectory would collide with one of these
# cannot be published.
RESERVED=(.git .nojekyll CNAME)
for entry in docs/* docs/.[!.]*; do
  [[ -e "$entry" ]] && RESERVED+=("$(basename "$entry")")
done

# sanitize converts a branch name into a URL-friendly
# directory name by replacing every character other
# than letters, digits, '.', '_', and '-' with '-'.
function sanitize {
  echo "$1" | sed 's/[^A-Za-z0-9._-]/-/g'
}

function is_reserved {
  local name
  for name in "${RESERVED[@]}"; do
    [[ "$name" == "$1" ]] && return 0
  done
  return 1
}

# ORPHAN names the temporary local branch used when the remote
# gh-pages branch doesn't exist yet.
ORPHAN="deploy-docs-$$"

# checkout_pages populates the directory $1 with a worktree
# of the latest remote gh-pages branch. If the remote branch
# doesn't exist, an empty orphan branch is created instead.
function checkout_pages {
  if git ls-remote --exit-code --heads origin "$PAGES_BRANCH" >/dev/null; then
    git fetch --quiet --depth 1 origin "$PAGES_BRANCH"
    git worktree add --quiet --detach "$1" FETCH_HEAD
  else
    git worktree add --quiet --detach "$1"
    git -C "$1" switch --quiet --orphan "$ORPHAN"
  fi
}

function cleanup {
  git worktree remove --force "$SITE" 2>/dev/null || true
  git branch --quiet -D "$ORPHAN" 2>/dev/null || true
}

# apply_main replaces the root of the site with /docs. Branch
# subdirectories are kept as long as their branch still exists
# on the remote, so stale branch docs are pruned here as well.
function apply_main {
  local site="$1" keep=() branch entry name
  while read -r branch; do
    name="$(sanitize "${branch#refs/heads/}")"
    is_reserved "$name" || keep+=("$name")
  done < <(git ls-remote --heads origin | cut -f2)

  for entry in "$site"/* "$site"/.[!.]*; do
    [[ -e "$entry" ]] || continue
    name="$(basename "$entry")"
    [[ "$name" == ".git" ]] && continue
    if [[ " ${keep[*]:-} " != *" $name "* ]]; then
      rm -rf "$entry"
    fi
  done

  cp -R docs/. "$site/"
}

# apply_branch replaces the subdirectory for the branch $2
# with /docs, or deletes it if $3 is set.
function apply_branch {
  local site="$1" name="$2" remove="${3:-}"
  rm -rf "${site:?}/$name"
  if [[ -z "$remove" ]]; then
    mkdir "$site/$name"
    cp -R docs/. "$site/$name/"
  fi
}

REMOVE=""
if [[ "${1:-}" == "--remove" ]]; then
  REMOVE="x"
  shift
fi

if [[ $# -ne 1 || -z "$1" ]]; then
  echo "Usage: $0 [--remove] <branch>" >&2
  exit 1
fi
BRANCH="$1"
NAME="$(sanitize "$BRANCH")"

if [[ "$BRANCH" == "$PAGES_BRANCH" ]]; then
  echo "Nothing to do for the '$PAGES_BRANCH' branch."
  exit 0
fi

if [[ "$BRANCH" == "$MAIN_BRANCH" && -n "$REMOVE" ]]; then
  echo "ERR! Refusing to remove the docs for '$MAIN_BRANCH'." >&2
  exit 1
fi

if [[ "$BRANCH" != "$MAIN_BRANCH" ]] && is_reserved "$NAME"; then
  echo "ERR! Branch '$BRANCH' maps to the directory '$NAME' which" >&2
  echo "collides with a file of the main site: ${RESERVED[*]}" >&2
  exit 1
fi

SITE="$(mktemp -d "${TMPDIR:-/tmp}/deploy-docs.XXXXXX")/site"
trap cleanup EXIT

# Other branches may push to gh-pages between our fetch and our
# push. When that happens, start over from the latest gh-pages.
for ((i = 1; i <= ATTEMPTS; i++)); do
  cleanup
  checkout_pages "$SITE"

  if [[ "$BRANCH" == "$MAIN_BRANCH" ]]; then
    apply_main "$SITE"
  else
    apply_branch "$SITE" "$NAME" "$REMOVE"
  fi

  # Serve the files as-is rather than processing them with Jekyll.
  echo "$DOMAIN" > "$SITE/CNAME"
  touch "$SITE/.nojekyll"

  git -C "$SITE" add --all
  if git -C "$SITE" diff --cached --quiet; then
    echo "No changes to publish."
    exit 0
  fi

  MESSAGE="Deploy docs for '$BRANCH' from $(git rev-parse --short HEAD)"
  [[ -n "$REMOVE" ]] && MESSAGE="Remove docs for '$BRANCH'"

  git -C "$SITE" \
    -c user.name="github-actions[bot]" \
    -c user.email="41898282+github-actions[bot]@users.noreply.github.com" \
    commit --quiet --message "$MESSAGE"

  if git -C "$SITE" push --quiet origin "HEAD:refs/heads/$PAGES_BRANCH"; then
    if [[ "$BRANCH" == "$MAIN_BRANCH" ]]; then
      echo "Published https://$DOMAIN/"
    elif [[ -n "$REMOVE" ]]; then
      echo "Removed https://$DOMAIN/$NAME/"
    else
      echo "Published https://$DOMAIN/$NAME/"
    fi
    exit 0
  fi

  echo "Push rejected (attempt $i of $ATTEMPTS). Retrying..."
  sleep $((i * 2))
done

echo "ERR! Failed to push to '$PAGES_BRANCH' after $ATTEMPTS attempts." >&2
exit 1
