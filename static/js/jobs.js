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

// Initialize jobs SSE functionality when DOM is ready
document.addEventListener('DOMContentLoaded', function() {
    initJobsSSE();
});