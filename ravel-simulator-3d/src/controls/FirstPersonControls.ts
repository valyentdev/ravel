import * as THREE from 'three';

export class FirstPersonControls {
  private camera: THREE.PerspectiveCamera;
  private domElement: HTMLElement;
  private isLocked = false;

  private moveForward = false;
  private moveBackward = false;
  private moveLeft = false;
  private moveRight = false;
  private canJump = false;
  private isSprinting = false;

  private velocity = new THREE.Vector3();
  private direction = new THREE.Vector3();
  private euler = new THREE.Euler(0, 0, 0, 'YXZ');

  private readonly MOVE_SPEED = 20.0;
  private readonly SPRINT_MULTIPLIER = 2.0;
  private readonly JUMP_VELOCITY = 10.0;
  private readonly GRAVITY = 30.0;
  private readonly MOUSE_SENSITIVITY = 0.002;

  constructor(camera: THREE.PerspectiveCamera, domElement: HTMLElement) {
    this.camera = camera;
    this.domElement = domElement;

    this.setupPointerLock();
    this.setupKeyboardControls();
  }

  private setupPointerLock() {
    this.domElement.addEventListener('click', () => {
      this.domElement.requestPointerLock();
    });

    document.addEventListener('pointerlockchange', () => {
      this.isLocked = document.pointerLockElement === this.domElement;
    });

    document.addEventListener('mousemove', (event) => {
      if (!this.isLocked) return;

      const movementX = event.movementX || 0;
      const movementY = event.movementY || 0;

      this.euler.setFromQuaternion(this.camera.quaternion);
      this.euler.y -= movementX * this.MOUSE_SENSITIVITY;
      this.euler.x -= movementY * this.MOUSE_SENSITIVITY;

      // Clamp vertical rotation
      this.euler.x = Math.max(-Math.PI / 2, Math.min(Math.PI / 2, this.euler.x));

      this.camera.quaternion.setFromEuler(this.euler);
    });
  }

  private setupKeyboardControls() {
    const onKeyDown = (event: KeyboardEvent) => {
      switch (event.code) {
        case 'KeyW':
        case 'ArrowUp':
          this.moveForward = true;
          break;
        case 'KeyS':
        case 'ArrowDown':
          this.moveBackward = true;
          break;
        case 'KeyA':
        case 'ArrowLeft':
          this.moveLeft = true;
          break;
        case 'KeyD':
        case 'ArrowRight':
          this.moveRight = true;
          break;
        case 'Space':
          if (this.canJump) {
            this.velocity.y = this.JUMP_VELOCITY;
            this.canJump = false;
          }
          break;
        case 'ShiftLeft':
        case 'ShiftRight':
          this.isSprinting = true;
          break;
      }
    };

    const onKeyUp = (event: KeyboardEvent) => {
      switch (event.code) {
        case 'KeyW':
        case 'ArrowUp':
          this.moveForward = false;
          break;
        case 'KeyS':
        case 'ArrowDown':
          this.moveBackward = false;
          break;
        case 'KeyA':
        case 'ArrowLeft':
          this.moveLeft = false;
          break;
        case 'KeyD':
        case 'ArrowRight':
          this.moveRight = false;
          break;
        case 'ShiftLeft':
        case 'ShiftRight':
          this.isSprinting = false;
          break;
      }
    };

    document.addEventListener('keydown', onKeyDown);
    document.addEventListener('keyup', onKeyUp);
  }

  public update(delta: number) {
    if (!this.isLocked) return;

    // Apply gravity
    this.velocity.y -= this.GRAVITY * delta;

    // Ground collision (simple ground at y=2)
    if (this.camera.position.y <= 2) {
      this.velocity.y = 0;
      this.camera.position.y = 2;
      this.canJump = true;
    }

    // Calculate movement direction
    this.direction.z = Number(this.moveForward) - Number(this.moveBackward);
    this.direction.x = Number(this.moveRight) - Number(this.moveLeft);
    this.direction.normalize();

    // Apply movement
    const speed = this.MOVE_SPEED * (this.isSprinting ? this.SPRINT_MULTIPLIER : 1.0);

    if (this.moveForward || this.moveBackward) {
      this.velocity.z -= this.direction.z * speed * delta;
    }

    if (this.moveLeft || this.moveRight) {
      this.velocity.x -= this.direction.x * speed * delta;
    }

    // Apply friction
    this.velocity.x *= 0.9;
    this.velocity.z *= 0.9;

    // Move camera
    const moveVector = new THREE.Vector3();
    moveVector.setFromMatrixColumn(this.camera.matrix, 0); // Right
    moveVector.multiplyScalar(-this.velocity.x * delta);
    this.camera.position.add(moveVector);

    moveVector.setFromMatrixColumn(this.camera.matrix, 0); // Forward
    moveVector.crossVectors(this.camera.up, moveVector);
    moveVector.multiplyScalar(-this.velocity.z * delta);
    this.camera.position.add(moveVector);

    // Apply vertical velocity
    this.camera.position.y += this.velocity.y * delta;
  }

  public getIsLocked(): boolean {
    return this.isLocked;
  }

  public getCameraPosition(): THREE.Vector3 {
    return this.camera.position;
  }

  public getCameraDirection(): THREE.Vector3 {
    const direction = new THREE.Vector3();
    this.camera.getWorldDirection(direction);
    return direction;
  }
}
