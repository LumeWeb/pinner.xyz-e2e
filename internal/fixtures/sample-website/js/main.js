// Sample IPFS Website JavaScript
document.addEventListener('DOMContentLoaded', function() {
    console.log('Sample IPFS Website loaded successfully');
    
    // Add simple interaction
    const features = document.querySelectorAll('.feature');
    features.forEach(feature => {
        feature.addEventListener('click', function() {
            this.style.transform = 'scale(1.05)';
            setTimeout(() => {
                this.style.transform = 'scale(1)';
            }, 200);
        });
    });
    
    console.log(`Found ${features.length} feature sections`);
});