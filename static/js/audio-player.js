// Audio Player Behavior - playlist management and keyboard controls for audio pages

function initAudioPlayer() {
    const playerState = {
        current: null
    };
    
    function play(item) {
        if (!item) return;
        
        if (item !== playerState.current) {
            // Remove current class from all items
            document.querySelectorAll('.ready').forEach(el => el.classList.remove('current'));
            
            // Update title
            const titleEl = document.querySelector('.title');
            if (titleEl) {
                titleEl.textContent = item.textContent;
            }
            
            // Update audio source
            const audio = document.querySelector('audio');
            if (audio) {
                audio.src = `/data/audio/${item.dataset.id}.mp3`;
            }
            
            // Set new current
            item.classList.add('current');
            playerState.current = item;
        }
        
        const audio = document.querySelector('audio');
        if (audio && audio.readyState > 1 && audio.paused) {
            audio.play();
        }
    }
    
    function pause() {
        const audio = document.querySelector('audio');
        if (audio) {
            audio.pause();
        }
    }
    
    function next() {
        if (!playerState.current) return;
        
        const items = Array.from(document.querySelectorAll('.ready'));
        const currentIndex = items.indexOf(playerState.current);
        const nextIndex = (currentIndex + 1) % items.length;
        play(items[nextIndex]);
    }
    
    function prev() {
        if (!playerState.current) return;
        
        const items = Array.from(document.querySelectorAll('.ready'));
        const currentIndex = items.indexOf(playerState.current);
        const prevIndex = (currentIndex - 1 + items.length) % items.length;
        play(items[prevIndex]);
    }
    
    // Keyboard controls for audio player
    document.addEventListener('keydown', function(event) {
        const audio = document.querySelector('audio');
        if (!audio) return;
        
        // Don't interfere if user is typing in an input field
        if (event.target.tagName === 'INPUT' || event.target.tagName === 'TEXTAREA') {
            return;
        }
        
        switch(event.key) {
            case ' ':
            case 'k':
                event.preventDefault();
                if (audio.paused) {
                    if (playerState.current) {
                        play(playerState.current);
                    } else {
                        const firstReady = document.querySelector('.ready');
                        if (firstReady) play(firstReady);
                    }
                } else {
                    pause();
                }
                break;
            case 'ArrowLeft':
                if (!event.metaKey && !event.altKey) {
                    event.preventDefault();
                    audio.currentTime -= 5;
                }
                break;
            case 'ArrowRight':
                if (!event.metaKey && !event.altKey) {
                    event.preventDefault();
                    audio.currentTime += 5;
                }
                break;
            case 'j':
                event.preventDefault();
                audio.currentTime -= 10;
                break;
            case 'l':
                event.preventDefault();
                audio.currentTime += 10;
                break;
            case 'ArrowUp':
                event.preventDefault();
                if (audio.volume < 0.9) {
                    audio.volume += 0.1;
                } else if (audio.volume < 1) {
                    audio.volume = 1;
                }
                break;
            case 'ArrowDown':
                event.preventDefault();
                if (audio.volume > 0.1) {
                    audio.volume -= 0.1;
                } else if (audio.volume > 0) {
                    audio.volume = 0;
                }
                break;
            case 'm':
                event.preventDefault();
                audio.muted = !audio.muted;
                break;
            case 'n':
                event.preventDefault();
                next();
                break;
            case 'p':
                event.preventDefault();
                prev();
                break;
        }
    });
    
    // Initialize with first ready item
    const firstReady = document.querySelector('.ready');
    if (firstReady) {
        play(firstReady);
    }
    
    // Audio event handlers
    document.addEventListener('canplay', function(event) {
        if (event.target.tagName === 'AUDIO' && event.target.paused) {
            event.target.play();
        }
    });
    
    document.addEventListener('ended', function(event) {
        if (event.target.tagName === 'AUDIO') {
            next();
        }
    });
    
    // Control button handlers
    document.addEventListener('click', function(event) {
        if (event.target.classList.contains('play')) {
            if (playerState.current) {
                play(playerState.current);
            } else {
                const firstReady = document.querySelector('.ready');
                if (firstReady) play(firstReady);
            }
        } else if (event.target.classList.contains('pause')) {
            pause();
        } else if (event.target.classList.contains('next')) {
            next();
        } else if (event.target.classList.contains('ready')) {
            const listItem = event.target.closest('li');
            if (listItem) {
                play(listItem);
            }
        }
    });
}

// Initialize audio player functionality when DOM is ready
document.addEventListener('DOMContentLoaded', function() {
    initAudioPlayer();
});