#!/bin/bash
# Double-click this file to fix "damaged app" warning on macOS
# This removes the quarantine attribute that macOS adds to downloaded apps

echo "Fixing DST DS Panel permissions..."
echo ""

if [ -d "/Applications/DST DS Panel.app" ]; then
    xattr -cr "/Applications/DST DS Panel.app"
    echo "✓ Done! You can now open DST DS Panel from Applications."
else
    echo "✗ DST DS Panel.app not found in /Applications"
    echo "  Please drag the app to Applications first, then run this script."
fi

echo ""
echo "Press any key to close..."
read -n 1
