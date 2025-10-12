// Common functionality shared across all pages

// External Link Behavior - confirm before navigation
function initExternalLinks() {
    document.addEventListener('click', function(event) {
        const link = event.target.closest('a[href^="http"]');
        if (link && !link.href.includes(window.location.hostname)) {
            if (!confirm('Are you sure you want to navigate away from this page?')) {
                event.preventDefault();
            }
        }
    });
}

// Form helpers - select all/none checkboxes
function initFormHelpers() {
    document.addEventListener('click', function(event) {
        const button = event.target;
        
        // Handle "all" buttons
        if (button.dataset.action === 'select-all') {
            event.preventDefault();
            const className = button.dataset.target;
            const checkboxes = document.querySelectorAll(`.${className}`);
            checkboxes.forEach(cb => cb.checked = true);
        }
        
        // Handle "none" buttons  
        if (button.dataset.action === 'select-none') {
            event.preventDefault();
            const className = button.dataset.target;
            const checkboxes = document.querySelectorAll(`.${className}`);
            checkboxes.forEach(cb => cb.checked = false);
        }
    });
}

// Initialize common functionality when DOM is ready
document.addEventListener('DOMContentLoaded', function() {
    initExternalLinks();
    initFormHelpers();
});