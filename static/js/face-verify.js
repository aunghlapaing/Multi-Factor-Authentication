// Face Verification JavaScript
document.addEventListener('DOMContentLoaded', function() {
    const video = document.getElementById('video');
    const canvas = document.getElementById('canvas');
    const statusMessage = document.getElementById('status-message');
    const spinner = document.getElementById('spinner');
    const verifyBtn = document.getElementById('verifyBtn');
    const faceDataInput = document.getElementById('faceData');
    
    let model;
    let stream;
    let faceDetected = false;
    let verificationAttempts = 0;
    const maxVerificationAttempts = 3;
    
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
            statusMessage.textContent = 'Error accessing camera. Please make sure you have granted camera permissions.';
            statusMessage.style.color = 'var(--error-color)';
            spinner.style.display = 'none';
        }
    }
    
    // Load face detection model
    async function loadFaceDetectionModel() {
        try {
            model = await blazeface.load();
            console.log('Face detection model loaded');
            statusMessage.textContent = 'Face detection ready. Looking for your face...';
        } catch (error) {
            console.error('Error loading face detection model:', error);
            statusMessage.textContent = 'Error loading face detection model. Please try again later.';
            statusMessage.style.color = 'var(--error-color)';
            spinner.style.display = 'none';
        }
    }
    
    // Detect faces in the video stream
    async function detectFaces() {
        if (!model || !video.readyState) return;
        
        const predictions = await model.estimateFaces(video, false);
        
        if (predictions.length > 0) {
            // Face detected
            if (!faceDetected) {
                faceDetected = true;
                statusMessage.textContent = 'Face detected! Verifying...';
                captureFaceAndVerify();
            }
            return true;
        } else {
            // No face detected
            if (faceDetected) {
                faceDetected = false;
                statusMessage.textContent = 'Face lost. Please look at the camera.';
            }
            return false;
        }
    }
    
    // Capture face image and send for verification
    async function captureFaceAndVerify() {
        const context = canvas.getContext('2d');
        canvas.width = video.videoWidth;
        canvas.height = video.videoHeight;
        context.drawImage(video, 0, 0, canvas.width, canvas.height);
        
        // Ensure canvas is responsive
        canvas.style.maxWidth = '100%';
        canvas.style.height = 'auto';
        canvas.style.borderRadius = '8px';
        
        // Convert to base64
        const imageData = canvas.toDataURL('image/jpeg');
        faceDataInput.value = imageData;
        
        // Send for verification
        verificationAttempts++;
        
        try {
            const response = await fetch('/api/verify-face', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ face_data: imageData }),
            });
            
            const result = await response.json();
            
            if (result.success) {
                // Verification successful
                statusMessage.textContent = 'Face verified! Redirecting...';
                statusMessage.style.color = 'var(--success-color)';
                spinner.style.display = 'none';
                
                // Use the global function if available, otherwise fallback to direct redirect
                if (typeof window.onFaceVerificationSuccess === 'function') {
                    window.onFaceVerificationSuccess();
                } else if (result.redirect) {
                    window.location.href = result.redirect;
                } else {
                    window.location.href = '/home';
                }
            } else {
                // Verification failed
                if (verificationAttempts < maxVerificationAttempts) {
                    statusMessage.textContent = `Verification failed. Attempt ${verificationAttempts}/${maxVerificationAttempts}. Please try again.`;
                    faceDetected = false;
                    setTimeout(() => {
                        statusMessage.textContent = 'Looking for your face...';
                    }, 2000);
                } else {
                    statusMessage.textContent = 'Maximum verification attempts reached. Please try manual verification.';
                    statusMessage.style.color = 'var(--error-color)';
                    spinner.style.display = 'none';
                    verifyBtn.style.display = 'block';
                    
                    // Stop camera
                    if (stream) {
                        stream.getTracks().forEach(track => track.stop());
                    }
                    clearInterval(detectFacesInterval);
                }
            }
        } catch (error) {
            console.error('Error verifying face:', error);
            statusMessage.textContent = 'Error verifying face. Please try again later.';
            statusMessage.style.color = 'var(--error-color)';
            spinner.style.display = 'none';
            verifyBtn.style.display = 'block';
        }
    }
    
    // Initialize
    async function init() {
        await setupCamera();
        await loadFaceDetectionModel();
        
        // Check for faces periodically
        detectFacesInterval = setInterval(detectFaces, 100);
    }
    
    // Start the application
    init();
});
