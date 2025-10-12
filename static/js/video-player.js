// Video Player Behavior - keyboard controls and media functionality for video pages

function initVideoPlayer() {
    document.addEventListener('keydown', function(event) {
        const video = document.querySelector('video');
        if (!video) return;
        
        // Don't interfere if user is typing in an input field
        if (event.target.tagName === 'INPUT' || event.target.tagName === 'TEXTAREA') {
            return;
        }
        
        switch(event.key) {
            case ' ':
            case 'k':
                event.preventDefault();
                if (video.paused) {
                    video.play();
                } else {
                    video.pause();
                }
                break;
            case 'ArrowLeft':
                if (!event.metaKey && !event.altKey) {
                    event.preventDefault();
                    video.currentTime -= 5;
                }
                break;
            case 'ArrowRight':
                if (!event.metaKey && !event.altKey) {
                    event.preventDefault();
                    video.currentTime += 5;
                }
                break;
            case 'j':
                event.preventDefault();
                video.currentTime -= 10;
                break;
            case 'l':
                event.preventDefault();
                video.currentTime += 10;
                break;
            case 'ArrowUp':
                event.preventDefault();
                if (video.volume < 0.9) {
                    video.volume += 0.1;
                } else if (video.volume < 1) {
                    video.volume = 1;
                }
                break;
            case 'ArrowDown':
                event.preventDefault();
                if (video.volume > 0.1) {
                    video.volume -= 0.1;
                } else if (video.volume > 0) {
                    video.volume = 0;
                }
                break;
            case 'm':
                event.preventDefault();
                video.muted = !video.muted;
                break;
            case 'f':
                event.preventDefault();
                if (document.fullscreenElement) {
                    document.exitFullscreen();
                } else {
                    video.requestFullscreen();
                }
                break;
        }
    });
    
    // Auto-play next video in playlist
    document.addEventListener('ended', function(event) {
        if (event.target.tagName === 'VIDEO') {
            const nextButton = document.getElementById('play-next');
            if (nextButton) {
                nextButton.click();
            }
        }
    });
    
    // Auto-play video if marked for autoplay
    document.addEventListener('DOMContentLoaded', function() {
        const autoplayVideo = document.querySelector('video[data-autoplay="true"]');
        if (autoplayVideo) {
            setTimeout(() => {
                autoplayVideo.play();
            }, 500);
        }
        
        // Handle auto-play-next functionality
        const autoplayNext = document.querySelector('[data-autoplay-next="true"]');
        if (autoplayNext) {
            const nextButton = document.getElementById('play-next');
            if (nextButton) {
                setTimeout(() => {
                    nextButton.click();
                }, 500);
            }
        }
    });
}

// Initialize video player functionality when DOM is ready
document.addEventListener('DOMContentLoaded', function() {
    initVideoPlayer();
});