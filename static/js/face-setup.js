// Face Setup JavaScript
document.addEventListener('DOMContentLoaded', function() {
    const video = document.getElementById('video');
    const canvas = document.getElementById('canvas');
    const captureBtn = document.getElementById('captureBtn');
    const retakeBtn = document.getElementById('retakeBtn');
    const saveBtn = document.getElementById('saveBtn');
    const faceDataInput = document.getElementById('faceData');
    
    let model;
    let stream;
    let capturedImage = false;
    
    // Initialize face detection
    async function setupCamera() {
        try {
            stream = await navigator.mediaDevices.getUserMedia({
                video: {
                    facingMode: 'user',
                    width: { ideal: 640 },
                    height: { ideal: 480 }
                },
                audio: false
            });
            
            video.srcObject = stream;
            
            return new Promise((resolve) => {
                video.onloadedmetadata = () => {
                    resolve(video);
                };
            });
        } catch (error) {
            console.error('Error accessing camera:', error);
            alert('Error accessing camera. Please make sure you have granted camera permissions.');
        }
    }
    
    // Load face detection model
    async function loadFaceDetectionModel() {
        try {
            model = await blazeface.load();
            console.log('Face detection model loaded');
        } catch (error) {
            console.error('Error loading face detection model:', error);
            alert('Error loading face detection model. Please try again later.');
        }
    }
    
    // Detect faces in the video stream
    async function detectFaces() {
        if (!model || !video.readyState) return;
        
        const predictions = await model.estimateFaces(video, false);
        
        if (predictions.length > 0) {
            // Face detected
            captureBtn.disabled = false;
            return true;
        } else {
            // No face detected
            captureBtn.disabled = true;
            return false;
        }
    }
    
    // Capture face image
    function captureFace() {
        const context = canvas.getContext('2d');
        canvas.width = video.videoWidth;
        canvas.height = video.videoHeight;
        context.drawImage(video, 0, 0, canvas.width, canvas.height);
        
        // Convert to base64
        const imageData = canvas.toDataURL('image/jpeg');
        faceDataInput.value = imageData;
        
        // Update UI
        capturedImage = true;
        captureBtn.style.display = 'none';
        retakeBtn.style.display = 'inline-block';
        saveBtn.disabled = false;
        
        // Stop camera
        if (stream) {
            stream.getTracks().forEach(track => track.stop());
        }
        video.style.display = 'none';
        canvas.style.display = 'block';
        
        // Ensure canvas is responsive
        canvas.style.maxWidth = '100%';
        canvas.style.height = 'auto';
        canvas.style.borderRadius = '8px';
    }
    
    // Retake photo
    function retakeFace() {
        // Reset UI
        capturedImage = false;
        captureBtn.style.display = 'inline-block';
        retakeBtn.style.display = 'none';
        saveBtn.disabled = true;
        faceDataInput.value = '';
        
        // Restart camera
        setupCamera().then(() => {
            video.style.display = 'block';
            canvas.style.display = 'none';
            detectFacesInterval = setInterval(detectFaces, 100);
        });
    }
    
    // Initialize
    async function init() {
        await setupCamera();
        await loadFaceDetectionModel();
        
        // Check for faces periodically
        detectFacesInterval = setInterval(detectFaces, 100);
        
        // Event listeners
        captureBtn.addEventListener('click', captureFace);
        retakeBtn.addEventListener('click', retakeFace);
    }
    
    // Start the application
    init();
});
