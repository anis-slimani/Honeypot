# Dashboard Dark Mode Guide

## Features

The HoneyGuard dashboard now includes a beautiful dark mode with:
- 🌙 Smooth theme transitions
- 💾 Persistent theme preference (saved to browser localStorage)
- 🎨 Optimized color palette for both light and dark modes
- 🔘 Easy toggle button in the navbar

## How to Use

### Toggle Dark Mode

1. **Via Toggle Button**:
   - Look for the theme toggle button in the top-right corner of the navbar
   - Click the button to switch between light and dark mode
   - The icon will change:
     - ☀️ Sun icon = Light mode
     - 🌙 Moon icon = Dark mode

2. **Automatic Persistence**:
   - Your theme preference is automatically saved
   - When you return to the dashboard, your last choice is remembered
   - Works across browser sessions

### Keyboard Shortcut (Optional)
You can add a keyboard shortcut by modifying the JavaScript:
```javascript
// Add this to the dashboard.html <script> section
document.addEventListener('keydown', (e) => {
    if (e.ctrlKey && e.key === 'd') {
        e.preventDefault();
        toggleTheme();
    }
});
```
This enables `Ctrl+D` to toggle dark mode.

## Color Schemes

### Light Mode
- **Background**: Light gray (#f8fafc)
- **Cards**: White (#ffffff)
- **Text**: Dark slate (#1e293b)
- **Primary**: Blue (#2563eb)
- **Borders**: Light gray (#e2e8f0)

### Dark Mode
- **Background**: Very dark blue (#0f172a)
- **Cards**: Dark slate (#1e293b)
- **Text**: Light gray (#e2e8f0)
- **Primary**: Bright blue (#3b82f6)
- **Borders**: Slate (#334155)

## Technical Implementation

### CSS Variables
The theme system uses CSS custom properties for dynamic theming:

```css
:root {
    /* Light mode (default) */
    --bg-primary: #f8fafc;
    --card-bg: #ffffff;
    --text: #1e293b;
}

[data-theme="dark"] {
    /* Dark mode */
    --bg-primary: #0f172a;
    --card-bg: #1e293b;
    --text: #e2e8f0;
}
```

### LocalStorage
Theme preference is stored in browser localStorage:
- Key: `theme`
- Values: `light` or `dark`
- Location: Browser's localStorage for the domain

### Toggle Function
```javascript
function toggleTheme() {
    const html = document.documentElement;
    const currentTheme = html.getAttribute('data-theme');
    const newTheme = currentTheme === 'dark' ? 'light' : 'dark';

    html.setAttribute('data-theme', newTheme);
    localStorage.setItem('theme', newTheme);

    const icon = document.getElementById('theme-icon');
    icon.textContent = newTheme === 'dark' ? '🌙' : '☀️';
}
```

## Customization

### Change Default Theme
To set dark mode as default, edit the JavaScript:

```javascript
// Change this line:
const savedTheme = localStorage.getItem('theme') || 'light';

// To:
const savedTheme = localStorage.getItem('theme') || 'dark';
```

### Customize Colors
Edit the CSS variables in `templates/dashboard.html`:

```css
[data-theme="dark"] {
    --primary: #3b82f6;        /* Change primary color */
    --bg-primary: #0f172a;     /* Change background */
    --card-bg: #1e293b;        /* Change card background */
    --text: #e2e8f0;           /* Change text color */
    /* Add more customizations */
}
```

### Add System Preference Detection
To automatically use the user's OS theme preference:

```javascript
// Add this before the theme loader
const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
const savedTheme = localStorage.getItem('theme') || (prefersDark ? 'dark' : 'light');
```

## Browser Compatibility

Fully compatible with:
- ✅ Chrome 88+
- ✅ Firefox 85+
- ✅ Safari 14+
- ✅ Edge 88+
- ✅ Opera 74+

Uses:
- CSS Custom Properties (CSS Variables)
- localStorage API
- data-attributes

## Screenshots

### Light Mode
```
┌─────────────────────────────────────────┐
│ 🛡️ HoneyGuard™          ☀️ ←Toggle    │
├─────────────────────────────────────────┤
│                                         │
│  [Stats Cards - White Background]      │
│                                         │
│  [Tabs - Light Blue Selected]          │
│                                         │
│  [Tables - White with Gray Headers]    │
│                                         │
└─────────────────────────────────────────┘
```

### Dark Mode
```
┌─────────────────────────────────────────┐
│ 🛡️ HoneyGuard™          🌙 ←Toggle    │
├─────────────────────────────────────────┤
│                                         │
│  [Stats Cards - Dark Gray Background]  │
│                                         │
│  [Tabs - Bright Blue Selected]         │
│                                         │
│  [Tables - Dark with Darker Headers]   │
│                                         │
└─────────────────────────────────────────┘
```

## Troubleshooting

### Theme doesn't persist
**Problem**: Theme resets to light mode on page refresh

**Solution**:
1. Check browser console for localStorage errors
2. Ensure cookies/storage is enabled
3. Try clearing localStorage: `localStorage.clear()`
4. Verify JavaScript is enabled

### Toggle button not visible
**Problem**: Can't see the toggle button

**Solution**:
1. On mobile, some nav items hide to save space
2. Check screen width - button appears on all sizes
3. Verify HTML was updated correctly
4. Check browser dev tools for CSS errors

### Colors look wrong
**Problem**: Some elements don't change color

**Solution**:
1. Hard refresh the page: `Ctrl+Shift+R` (or `Cmd+Shift+R` on Mac)
2. Clear browser cache
3. Check for conflicting CSS rules in browser dev tools

### Toggle doesn't work
**Problem**: Clicking toggle does nothing

**Solution**:
1. Check browser console for JavaScript errors
2. Verify `toggleTheme()` function is defined
3. Ensure `theme-icon` element ID exists
4. Check onclick handler is attached

## Performance

- **Zero performance impact** - Uses CSS variables (hardware accelerated)
- **Instant switching** - No page reload required
- **Smooth transitions** - 0.3s ease animation
- **Lightweight** - No external dependencies

## Accessibility

- ✅ Proper contrast ratios (WCAG AA compliant)
- ✅ Works with screen readers
- ✅ Keyboard accessible toggle button
- ✅ `aria-label` on toggle button
- ✅ Respects reduced motion preferences (optional)

To add reduced motion support:
```css
@media (prefers-reduced-motion: reduce) {
    * {
        transition-duration: 0.01ms !important;
    }
}
```

## Future Enhancements

Possible additions:
- 🌈 Additional color themes (blue, purple, green)
- ⏰ Automatic theme based on time of day
- 🎨 Theme customizer panel
- 📱 Mobile-optimized toggle placement
- ⚙️ Per-user theme settings (requires backend)

## Testing

### Manual Testing Checklist
- [ ] Toggle between light and dark mode
- [ ] Refresh page - theme persists
- [ ] Check all dashboard tabs in both modes
- [ ] Verify table readability in dark mode
- [ ] Test on mobile devices
- [ ] Check with browser DevTools color picker
- [ ] Verify contrast ratios

### Browser Developer Tools
```javascript
// Test in browser console:

// Get current theme
console.log(document.documentElement.getAttribute('data-theme'));

// Force dark mode
document.documentElement.setAttribute('data-theme', 'dark');

// Force light mode
document.documentElement.setAttribute('data-theme', 'light');

// Check saved preference
console.log(localStorage.getItem('theme'));

// Clear saved preference
localStorage.removeItem('theme');
```

## Credits

Dark mode implementation using:
- CSS Custom Properties (CSS Variables)
- Vanilla JavaScript (no frameworks)
- LocalStorage API
- Modern CSS transitions

Designed for the HoneyGuard™ Security Platform.
