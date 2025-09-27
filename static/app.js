// Standard JavaScript replacement for htmx and hyperscript functionality

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

// Video Player Behavior - keyboard controls and media functionality
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

// Audio Player Behavior - playlist management and controls
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

// Jobs Page SSE - Server-Sent Events for real-time updates
function initJobsSSE() {
    const jobsTable = document.querySelector('table[data-jobs-sse]');
    if (!jobsTable) return;
    
    const eventSource = new EventSource('/jobs/updates');
    
    eventSource.onmessage = function(event) {
        try {
            const data = JSON.parse(event.data);
            const jobId = 'job-' + data.id;
            const jobRow = document.getElementById(jobId);
            
            if (jobRow) {
                // Update status
                const statusCell = jobRow.querySelector('.job-status');
                if (statusCell) {
                    statusCell.textContent = data.status.charAt(0).toUpperCase() + data.status.slice(1);
                }
                
                // Update progress
                const progressCell = jobRow.querySelector('.job-progress');
                if (progressCell && data.progress !== null) {
                    const queueName = jobRow.cells[1].textContent;
                    if (queueName === 'video_download' || queueName === 'video_transcode') {
                        progressCell.innerHTML = `
                            <div class="progress-container">
                                <div class="progress-bar" style="width: ${data.progress}%;">
                                    ${data.progress}%
                                </div>
                            </div>
                        `;
                    } else {
                        progressCell.textContent = data.progress + '%';
                    }
                }
                
                // Update row classes
                jobRow.classList.remove('job-running', 'job-finished');
                if (data.status === 'running') {
                    jobRow.classList.add('job-running');
                } else if (data.status === 'finished') {
                    jobRow.classList.add('job-finished');
                }
            }
        } catch (error) {
            console.error('Error parsing SSE data:', error);
        }
    };
    
    eventSource.onerror = function(event) {
        console.error('SSE connection error:', event);
    };
    
    // Clean up on page unload
    window.addEventListener('beforeunload', function() {
        eventSource.close();
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

// Initialize all functionality when DOM is ready
document.addEventListener('DOMContentLoaded', function() {
    initExternalLinks();
    initVideoPlayer();
    initAudioPlayer();
    initJobsSSE();
    initFormHelpers();
});