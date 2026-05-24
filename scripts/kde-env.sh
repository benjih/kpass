# Source from devbox init_hook or before go build / go run.
# KDE Frameworks 6.6 in nixpkgs are built against Qt 6.7; miqt must use the same Qt at
# link time and runtime (do not add qt6.full — it is Qt 6.6 and breaks KDE QML plugins).

profile_store_paths() {
  if [ -e "${DEVBOX_PROJECT_ROOT:-.}/.devbox/nix/profile/default" ]; then
    nix-store -qR "${DEVBOX_PROJECT_ROOT:-.}/.devbox/nix/profile/default" 2>/dev/null
  fi
}

find_store() {
  local pattern="$1"
  profile_store_paths | grep -E "$pattern" | head -1
}

QTBASE="$(find_store 'qtbase-6\.7\.[0-9]+$')"
QTBASE_DEV="$(find_store 'qtbase-6\.7\.[0-9]+-dev$')"
QTDECL="$(find_store 'qtdeclarative-6\.7\.[0-9]+$')"

qt_pkgconfig_dirs() {
  local dirs=""
  for root in "$QTBASE" "$QTBASE_DEV" "$QTDECL"; do
    if [ -n "$root" ] && [ -d "$root/lib/pkgconfig" ]; then
      if [ -n "$dirs" ]; then
        dirs="$dirs:$root/lib/pkgconfig"
      else
        dirs="$root/lib/pkgconfig"
      fi
    fi
  done
  printf '%s' "$dirs"
}

PC_DIRS="$(qt_pkgconfig_dirs)"
if [ -n "$PC_DIRS" ]; then
  export PKG_CONFIG_PATH="$PC_DIRS"
fi

if [ -n "$QTBASE" ] && [ -d "$QTBASE/lib" ]; then
  export LD_LIBRARY_PATH="$QTBASE/lib${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"
  export PATH="$QTBASE/bin:$PATH"
fi
if [ -n "$QTDECL" ] && [ -d "$QTDECL/lib" ]; then
  export LD_LIBRARY_PATH="$QTDECL/lib:$LD_LIBRARY_PATH"
fi

kde_qml_paths() {
  local paths=""
  if [ -d "${DEVBOX_PROJECT_ROOT:-.}/.devbox/nix/profile/default/lib/qt-6/qml" ]; then
    paths="${DEVBOX_PROJECT_ROOT:-.}/.devbox/nix/profile/default/lib/qt-6/qml"
  fi
  if command -v nix-store >/dev/null 2>&1; then
    while IFS= read -r store; do
      if [ -d "$store/lib/qt-6/qml" ]; then
        if [ -n "$paths" ]; then
          paths="$paths:$store/lib/qt-6/qml"
        else
          paths="$store/lib/qt-6/qml"
        fi
      fi
    done < <(profile_store_paths)
  fi
  printf '%s' "$paths"
}

kde_plugin_paths() {
  local paths=""
  append_if_exists() {
    local p="$1"
    if [ -d "$p" ]; then
      case ":${paths}:" in
        *:"$p":*) ;;
        *) paths="${paths:+$paths:}$p" ;;
      esac
    fi
  }

  append_if_exists "${DEVBOX_PROJECT_ROOT:-.}/.devbox/nix/profile/default/lib/qt-6/plugins"
  if [ -n "$QTBASE" ]; then
    append_if_exists "$QTBASE/lib/qt-6/plugins"
  fi
  if command -v nix-store >/dev/null 2>&1; then
    while IFS= read -r store; do
      append_if_exists "$store/lib/qt-6/plugins"
    done < <(profile_store_paths)
  fi
  printf '%s' "$paths"
}

export QML2_IMPORT_PATH="$(kde_qml_paths)"
export QT_QUICK_CONTROLS_STYLE="${QT_QUICK_CONTROLS_STYLE:-org.kde.desktop}"
export XDG_ICON_THEME="${XDG_ICON_THEME:-breeze}"

# Breeze icons + KDE platform theme (frameworkintegration) for native fonts/colors.
PROFILE="${DEVBOX_PROJECT_ROOT:-.}/.devbox/nix/profile/default"
if [ -d "$PROFILE/share" ]; then
  case ":${XDG_DATA_DIRS:-}:" in
    *:"$PROFILE/share":*) ;;
    *) export XDG_DATA_DIRS="$PROFILE/share${XDG_DATA_DIRS:+:$XDG_DATA_DIRS}" ;;
  esac
  export KPASS_ICON_PROFILE="$PROFILE/share"
fi

# KIconEnginePlugin (kiconthemes) and FrameworkIntegrationPlugin live in the nix
# closure but not under the profile's sparse plugins/ tree; include all plugin dirs.
KDE_PLUGINS="$(kde_plugin_paths)"
if [ -n "$KDE_PLUGINS" ]; then
  export QT_PLUGIN_PATH="$KDE_PLUGINS${QT_PLUGIN_PATH:+:$QT_PLUGIN_PATH}"
fi

# Prefer KDE platform integration when available (matches Plasma apps).
if [ -z "${QT_QPA_PLATFORMTHEME:-}" ] && [ -n "$(find_store 'frameworkintegration-6\.[0-9.]+$')" ]; then
  export QT_QPA_PLATFORMTHEME=kde
fi

# miqt pulls many Qt headers into separate CGO objects; allow duplicate weak Qt template symbols at link.
export CGO_LDFLAGS="${CGO_LDFLAGS} -Wl,--allow-multiple-definition"
