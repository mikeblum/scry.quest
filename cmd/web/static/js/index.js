// Constants
const REFOCUS_DELAY_MS = 200;
const MOCK_RESPONSE_DELAY_MS = 300;

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
    container.scrollTop = container.scrollHeight;
  }
}

function showLoadingState() {
  const submitBtn = document.getElementById('chat-submit-btn');
  const icon = document.getElementById('submit-icon');
  const userInput = document.getElementById('user-input');

  if (submitBtn) submitBtn.disabled = true;
  if (icon) {
    icon.textContent = 'refresh';
    icon.classList.add('animate-spin');
  }
  if (userInput) userInput.disabled = true;
}

function hideLoadingState() {
  const submitBtn = document.getElementById('chat-submit-btn');
  const icon = document.getElementById('submit-icon');
  const userInput = document.getElementById('user-input');

  if (submitBtn) submitBtn.disabled = false;
  if (icon) {
    icon.textContent = 'send';
    icon.classList.remove('animate-spin');
  }
  if (userInput) userInput.disabled = false;
}

function handleMockSubmit(event) {
  event.preventDefault();

  const userInput = document.getElementById('user-input');
  const messagesContainer = document.getElementById('messages');

  if (!userInput.value.trim()) {
    return false;
  }

  const userMessage = userInput.value.trim();

  // Show loading state
  showLoadingState();
  hideWelcomeScreen();

  // Add user message
  const userBubbleHtml = `
    <div class="mb-6 flex justify-end chat-message">
      <div class="max-w-3xl bg-crystal-600 text-white rounded-lg px-4 py-3">
        <div class="whitespace-pre-wrap">${userMessage}</div>
      </div>
    </div>
  `;

  if (messagesContainer) {
    messagesContainer.insertAdjacentHTML('beforeend', userBubbleHtml);
  }

  // Clear input
  userInput.value = '';
  userInput.style.height = 'auto';

  // Scroll to bottom
  setTimeout(scrollToBottom, 100);

  // Mock AI response after delay
  setTimeout(() => {
    const aiResponses = [
      "That's an interesting question about D&D 5e! Let me help you with that.",
      "Here's what you need to know about that spell/monster/rule...",
      "Great question! In D&D 5e, this works differently than you might expect.",
      "I can definitely help you understand that mechanic better.",
      "That's a common question among D&D players. Here's the answer..."
    ];

    const randomResponse = aiResponses[Math.floor(Math.random() * aiResponses.length)];

    const aiBubbleHtml = `
      <div class="mb-6 chat-message">
        <div class="flex items-start space-x-3">
          <div class="w-8 h-8 flex items-center justify-center flex-shrink-0">
            <span class="text-2xl">🔮</span>
          </div>
          <div class="flex-1 min-w-0">
            <div class="text-primary whitespace-pre-wrap leading-relaxed">
              ${randomResponse}
            </div>
          </div>
        </div>
      </div>
    `;

    if (messagesContainer) {
      messagesContainer.insertAdjacentHTML('beforeend', aiBubbleHtml);
    }

    // Hide loading state and scroll
    hideLoadingState();
    setTimeout(scrollToBottom, 300);

    // Refocus on the input after response
    setTimeout(() => {
      const input = document.getElementById('user-input');
      if (input) {
        input.focus();
      }
    }, REFOCUS_DELAY_MS);

  }, MOCK_RESPONSE_DELAY_MS);

  return false;
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