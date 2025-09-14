// Theme management
function initTheme() {
  const theme = localStorage.getItem('theme') ||
    (window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark');
  document.documentElement.className = theme;
  updateThemeIcon(theme);
}

function toggleTheme() {
  const currentTheme = document.documentElement.className;
  const newTheme = currentTheme === 'light' ? 'dark' : 'light';
  document.documentElement.className = newTheme;
  localStorage.setItem('theme', newTheme);
  updateThemeIcon(newTheme);
}

function updateThemeIcon(theme) {
  const icon = document.getElementById('theme-icon');
  if (icon) {
    icon.textContent = theme === 'light' ? 'dark_mode' : 'light_mode';
  }
}

// Initialize theme on load
document.addEventListener('DOMContentLoaded', initTheme);

// Listen for system theme changes
window.matchMedia('(prefers-color-scheme: light)').addEventListener('change', (e) => {
  if (!localStorage.getItem('theme')) {
    const theme = e.matches ? 'light' : 'dark';
    document.documentElement.className = theme;
    updateThemeIcon(theme);
  }
});


// Chat functionality
function hideWelcomeScreen() {
  const welcomeScreen = document.getElementById('welcome-screen');
  const welcomeSection = document.getElementById('welcome-section');
  if (welcomeScreen) {
    welcomeScreen.style.display = 'none';
  }
  if (welcomeSection) {
    welcomeSection.style.display = 'none';
  }
}

function scrollToBottom() {
  const container = document.getElementById('messages-scroll-container');
  if (container) {
    setTimeout(() => {
      container.scrollTop = container.scrollHeight;
    }, 100);
  }
}

// Auto-expand textarea
function autoExpandTextarea(textarea) {
  textarea.style.height = 'auto';
  textarea.style.height = Math.min(textarea.scrollHeight, 120) + 'px';
}

// Sidebar functionality
function toggleSidebar() {
  const sidebar = document.getElementById('sidebar');
  const overlay = document.getElementById('mobile-overlay');
  const openBtn = document.getElementById('open-sidebar-btn');

  if (sidebar) {
    const isOpen = !sidebar.classList.contains('-translate-x-full');

    if (isOpen) {
      // Closing sidebar
      sidebar.classList.add('-translate-x-full');
      if (overlay) overlay.classList.add('hidden');
      // On mobile, show the open button when sidebar is closed
      if (openBtn && window.innerWidth < 768) {
        openBtn.classList.remove('hidden');
      }
    } else {
      // Opening sidebar
      sidebar.classList.remove('-translate-x-full');
      if (overlay) overlay.classList.remove('hidden');
      // Hide the open button when sidebar is open
      if (openBtn) openBtn.classList.add('hidden');
    }
  }
}



// Initialize on load
document.addEventListener('DOMContentLoaded', () => {
  initTheme();

  // Setup textarea auto-expand and submit button state
  const userInput = document.getElementById('user-input');
  const submitButton = document.querySelector('#chat-form button[type="submit"]');
  const openBtn = document.getElementById('open-sidebar-btn');

  // Set initial open button visibility based on screen size
  if (openBtn) {
    if (window.innerWidth < 768) {
      openBtn.classList.remove('hidden');
    } else {
      openBtn.classList.add('hidden');
    }
  }

  if (userInput && submitButton) {
    // Auto-expand textarea
    userInput.addEventListener('input', () => {
      autoExpandTextarea(userInput);
      // Update submit button state
      submitButton.disabled = !userInput.value.trim();
    });

    // Initial button state
    submitButton.disabled = !userInput.value.trim();

    // Focus on chat input on larger screens
    if (window.innerWidth >= 768) {
      userInput.focus();
    }
  }

  // Handle window resize behavior
  window.addEventListener('resize', () => {
    const sidebar = document.getElementById('sidebar');
    const overlay = document.getElementById('mobile-overlay');
    const openBtn = document.getElementById('open-sidebar-btn');

    if (window.innerWidth >= 768) {
      // Tablet/Desktop: ensure sidebar is visible, overlay hidden, open button hidden
      if (sidebar) sidebar.classList.remove('-translate-x-full');
      if (overlay) overlay.classList.add('hidden');
      if (openBtn) openBtn.classList.add('hidden');
    } else {
      // Mobile: close sidebar by default, show open button
      if (sidebar) sidebar.classList.add('-translate-x-full');
      if (overlay) overlay.classList.add('hidden');
      if (openBtn) openBtn.classList.remove('hidden');
    }
  });
});