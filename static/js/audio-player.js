// Audio Player Behavior - playlist management and keyboard controls for audio pages

function initAudioPlayer() {
    const playerState = {
        current: null
    };

    const audio = document.querySelector('audio');
    if (!audio) {
        console.error("Audio element not found. Cannot initialize player.");
        return; // Cannot initialize player without audio element
    }

    // Attach ended listener directly to the audio element for reliable track progression
    audio.addEventListener('ended', function() {
        next();
    });


    function play(item, shouldAutoplay = true) {
        if (!item) return;

        // Remove current class from all items
        document.querySelectorAll('.ready').forEach(el => el.classList.remove('current'));

        // Update title
        const titleEl = document.querySelector('.title');
        if (titleEl) {
            titleEl.textContent = item.textContent;
        }

        // Store the current paused state before changing source
         const wasPaused = audio.paused;

        // Update audio source
        audio.src = `/data/audio/${item.dataset.id}.mp3`;

        // Set new current
        item.classList.add('current');
        playerState.current = item;

        // Explicitly attempt to play if autoplay is requested or if it was playing before
        if (shouldAutoplay || !wasPaused) {
            // Use a canplay listener as a fallback if immediate play fails
            const tempCanplayListener = function() {
                // Check if it's still paused before playing
                 if (audio.paused) {
                    audio.play().catch(error => {
                        console.error("Autoplay prevented on canplay:", error);
                    });
                 }
                audio.removeEventListener('canplay', tempCanplayListener);
            };
            audio.addEventListener('canplay', tempCanplayListener);

            // Attempt to play immediately (might be blocked by autoplay policies)
             audio.play().catch(error => {
                 console.warn("Initial play call failed:", error);
                 // The canplay listener will attempt to play again when ready
             });
        }
    }

    function pause() {
        if (audio) {
            audio.pause();
        }
    }

    function next() {
        if (!playerState.current) return;

        const items = Array.from(document.querySelectorAll('.ready'));
        const currentIndex = items.indexOf(playerState.current);
        const nextIndex = (currentIndex + 1) % items.length;
        play(items[nextIndex], true); // Autoplay the next track
    }

    function prev() {
        if (!playerState.current) return;

        const items = Array.from(document.querySelectorAll('.ready'));
        const currentIndex = items.indexOf(playerState.current);
        const prevIndex = (currentIndex - 1 + items.length) % items.length;
        play(items[prevIndex], true); // Autoplay the previous track
    }

    // Keyboard controls for audio player
    document.addEventListener('keydown', function(event) {
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
                        audio.play().catch(error => console.error("Play failed:", error));
                    } else {
                        const firstReady = document.querySelector('.ready');
                        if (firstReady) {
                            play(firstReady, true); // Set source and autoplay
                        }
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
        play(firstReady, true); // Autoplay the first track on load
    }

    // Control button handlers
    document.addEventListener('click', function(event) {
        if (!audio) return;

        if (event.target.classList.contains('play')) {
            if (playerState.current) {
                audio.play().catch(error => console.error("Play failed:", error));
            } else {
                const firstReady = document.querySelector('.ready');
                if (firstReady) {
                    play(firstReady, true); // Set source and autoplay
                }
            }
        } else if (event.target.classList.contains('pause')) {
            pause();
        } else if (event.target.classList.contains('next')) {
            next();
        } else if (event.target.classList.contains('ready')) {
            const listItem = event.target.closest('li');
            if (listItem) {
                play(listItem, true); // Set source and autoplay
            }
        }
    });
}

// Initialize audio player functionality when DOM is ready
document.addEventListener('DOMContentLoaded', function() {
    initAudioPlayer();
});

