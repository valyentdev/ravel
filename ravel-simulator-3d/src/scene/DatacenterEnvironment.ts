import * as THREE from 'three';
import type { Region } from '../types';

export class DatacenterEnvironment {
  private group: THREE.Group;
  private regions: Region[];

  constructor(regions: Region[]) {
    this.group = new THREE.Group();
    this.regions = regions;
    this.build();
  }

  private build() {
    // Create industrial floor
    this.createIndustrialFloor();

    // Create datacenter structure for each region
    this.regions.forEach((region) => {
      this.createRegionDatacenter(region);
    });

    // Add environmental elements
    this.createCeilingStructure();
    this.createOverheadLighting();
    this.createCableTraySystem();
  }

  private createIndustrialFloor() {
    // Main datacenter floor - larger area
    const floorGeometry = new THREE.PlaneGeometry(300, 300);

    // Create industrial floor material with grid pattern
    const canvas = document.createElement('canvas');
    canvas.width = 512;
    canvas.height = 512;
    const ctx = canvas.getContext('2d')!;

    // Base dark concrete color
    ctx.fillStyle = '#0a0a0f';
    ctx.fillRect(0, 0, 512, 512);

    // Add grid lines (like datacenter floor tiles)
    ctx.strokeStyle = '#1a1a2e';
    ctx.lineWidth = 2;

    const tileSize = 64;
    for (let x = 0; x <= 512; x += tileSize) {
      ctx.beginPath();
      ctx.moveTo(x, 0);
      ctx.lineTo(x, 512);
      ctx.stroke();
    }
    for (let y = 0; y <= 512; y += tileSize) {
      ctx.beginPath();
      ctx.moveTo(0, y);
      ctx.lineTo(512, y);
      ctx.stroke();
    }

    // Add some wear/dirt texture
    for (let i = 0; i < 200; i++) {
      ctx.fillStyle = `rgba(20, 20, 30, ${Math.random() * 0.3})`;
      const x = Math.random() * 512;
      const y = Math.random() * 512;
      const size = Math.random() * 20 + 5;
      ctx.fillRect(x, y, size, size);
    }

    const floorTexture = new THREE.CanvasTexture(canvas);
    floorTexture.wrapS = THREE.RepeatWrapping;
    floorTexture.wrapT = THREE.RepeatWrapping;
    floorTexture.repeat.set(4, 4);

    const floorMaterial = new THREE.MeshStandardMaterial({
      map: floorTexture,
      roughness: 0.8,
      metalness: 0.2,
    });

    const floor = new THREE.Mesh(floorGeometry, floorMaterial);
    floor.rotation.x = -Math.PI / 2;
    floor.position.y = -0.1;
    floor.receiveShadow = true;
    this.group.add(floor);

    // Add raised floor panels appearance
    this.createRaisedFloorPanels();
  }

  private createRaisedFloorPanels() {
    // Add some elevated platforms under server areas
    const platformGeometry = new THREE.BoxGeometry(15, 0.3, 15);
    const platformMaterial = new THREE.MeshStandardMaterial({
      color: 0x1a1a2e,
      roughness: 0.7,
      metalness: 0.3,
    });

    this.regions.forEach((region) => {
      if (region.nodes.length === 0) return;

      // Calculate region center
      const centerX = region.nodes.reduce((sum, n) => sum + n.position.x, 0) / region.nodes.length;
      const centerZ = region.nodes.reduce((sum, n) => sum + n.position.z, 0) / region.nodes.length;

      const platform = new THREE.Mesh(platformGeometry, platformMaterial);
      platform.position.set(centerX, 0.05, centerZ);
      platform.receiveShadow = true;
      platform.castShadow = true;
      this.group.add(platform);

      // Add platform edge lights
      this.addPlatformEdgeLights(centerX, centerZ);
    });
  }

  private addPlatformEdgeLights(x: number, z: number) {
    const positions = [
      { x: x - 7, z: z - 7 },
      { x: x + 7, z: z - 7 },
      { x: x - 7, z: z + 7 },
      { x: x + 7, z: z + 7 },
    ];

    positions.forEach((pos) => {
      const lightGeometry = new THREE.CylinderGeometry(0.1, 0.1, 0.5, 8);
      const lightMaterial = new THREE.MeshBasicMaterial({
        color: 0x00ff88,
        emissive: 0x00ff88,
        emissiveIntensity: 1,
      });

      const light = new THREE.Mesh(lightGeometry, lightMaterial);
      light.position.set(pos.x, 0.3, pos.z);
      this.group.add(light);

      // Add point light for ambient glow
      const pointLight = new THREE.PointLight(0x00ff88, 0.5, 5);
      pointLight.position.set(pos.x, 0.5, pos.z);
      this.group.add(pointLight);
    });
  }

  private createRegionDatacenter(region: Region) {
    if (region.nodes.length === 0) return;

    // Calculate region bounds
    const centerX = region.nodes.reduce((sum, n) => sum + n.position.x, 0) / region.nodes.length;
    const centerZ = region.nodes.reduce((sum, n) => sum + n.position.z, 0) / region.nodes.length;

    // Create region boundary with industrial barriers
    this.createRegionBoundary(centerX, centerZ, region.name);

    // Add cooling units
    this.createCoolingUnit(centerX + 10, 0, centerZ);
    this.createCoolingUnit(centerX - 10, 0, centerZ);

    // Add cable management pillars
    this.createCablePillar(centerX, 0, centerZ + 10);
    this.createCablePillar(centerX, 0, centerZ - 10);
  }

  private createRegionBoundary(x: number, z: number, name: string) {
    // Create low industrial barriers around region
    const barrierGeometry = new THREE.BoxGeometry(20, 0.5, 0.2);
    const barrierMaterial = new THREE.MeshStandardMaterial({
      color: 0xffaa00,
      roughness: 0.8,
      metalness: 0.4,
      emissive: 0xffaa00,
      emissiveIntensity: 0.2,
    });

    const positions = [
      { x, z: z - 10, rotY: 0 },
      { x, z: z + 10, rotY: 0 },
      { x: x - 10, z, rotY: Math.PI / 2 },
      { x: x + 10, z, rotY: Math.PI / 2 },
    ];

    positions.forEach((pos) => {
      const barrier = new THREE.Mesh(barrierGeometry, barrierMaterial);
      barrier.position.set(pos.x, 0.25, pos.z);
      barrier.rotation.y = pos.rotY;
      barrier.castShadow = true;
      this.group.add(barrier);

      // Add hazard stripes
      const stripeGeometry = new THREE.PlaneGeometry(20, 0.3);
      const stripeMaterial = new THREE.MeshBasicMaterial({
        color: 0xffff00,
        transparent: true,
        opacity: 0.6,
      });
      const stripe = new THREE.Mesh(stripeGeometry, stripeMaterial);
      stripe.position.set(pos.x, 0.51, pos.z);
      stripe.rotation.x = -Math.PI / 2;
      stripe.rotation.z = pos.rotY;
      this.group.add(stripe);
    });

    // Add region name label
    this.createRegionLabel(x, z + 12, name);
  }

  private createRegionLabel(x: number, z: number, name: string) {
    const canvas = document.createElement('canvas');
    canvas.width = 512;
    canvas.height = 128;
    const ctx = canvas.getContext('2d')!;

    // Background
    ctx.fillStyle = '#000000';
    ctx.fillRect(0, 0, 512, 128);

    // Border
    ctx.strokeStyle = '#00ff88';
    ctx.lineWidth = 4;
    ctx.strokeRect(5, 5, 502, 118);

    // Text
    ctx.fillStyle = '#00ff88';
    ctx.font = 'bold 48px monospace';
    ctx.textAlign = 'center';
    ctx.fillText(name, 256, 80);

    const texture = new THREE.CanvasTexture(canvas);
    const material = new THREE.MeshBasicMaterial({
      map: texture,
      transparent: true,
      side: THREE.DoubleSide,
    });

    const geometry = new THREE.PlaneGeometry(8, 2);
    const label = new THREE.Mesh(geometry, material);
    label.position.set(x, 2, z);
    this.group.add(label);
  }

  private createCoolingUnit(x: number, y: number, z: number) {
    const group = new THREE.Group();

    // Main cooling unit body
    const bodyGeometry = new THREE.BoxGeometry(2, 4, 1.5);
    const bodyMaterial = new THREE.MeshStandardMaterial({
      color: 0x2a2a3a,
      roughness: 0.6,
      metalness: 0.5,
    });
    const body = new THREE.Mesh(bodyGeometry, bodyMaterial);
    body.position.y = 2;
    body.castShadow = true;
    group.add(body);

    // Vent grill
    for (let i = 0; i < 8; i++) {
      const ventGeometry = new THREE.BoxGeometry(1.8, 0.1, 0.05);
      const ventMaterial = new THREE.MeshStandardMaterial({
        color: 0x444455,
        roughness: 0.8,
      });
      const vent = new THREE.Mesh(ventGeometry, ventMaterial);
      vent.position.set(0, 1 + i * 0.4, 0.76);
      group.add(vent);
    }

    // Status LED
    const ledGeometry = new THREE.SphereGeometry(0.1, 16, 16);
    const ledMaterial = new THREE.MeshBasicMaterial({
      color: 0x00ff00,
      emissive: 0x00ff00,
      emissiveIntensity: 1,
    });
    const led = new THREE.Mesh(ledGeometry, ledMaterial);
    led.position.set(0.7, 3.5, 0.76);
    group.add(led);

    // Ambient cooling light
    const coolLight = new THREE.PointLight(0x00aaff, 0.5, 8);
    coolLight.position.set(0, 2, 0);
    group.add(coolLight);

    group.position.set(x, y, z);
    this.group.add(group);
  }

  private createCablePillar(x: number, y: number, z: number) {
    const pillarGroup = new THREE.Group();

    // Main pillar
    const pillarGeometry = new THREE.CylinderGeometry(0.3, 0.3, 8, 16);
    const pillarMaterial = new THREE.MeshStandardMaterial({
      color: 0x333344,
      roughness: 0.7,
      metalness: 0.5,
    });
    const pillar = new THREE.Mesh(pillarGeometry, pillarMaterial);
    pillar.position.y = 4;
    pillar.castShadow = true;
    pillarGroup.add(pillar);

    // Cable holders
    for (let i = 1; i <= 6; i++) {
      const holderGeometry = new THREE.TorusGeometry(0.4, 0.05, 8, 16);
      const holderMaterial = new THREE.MeshStandardMaterial({
        color: 0x555566,
        roughness: 0.8,
        metalness: 0.3,
      });
      const holder = new THREE.Mesh(holderGeometry, holderMaterial);
      holder.position.y = i * 1.2;
      holder.rotation.x = Math.PI / 2;
      pillarGroup.add(holder);
    }

    pillarGroup.position.set(x, y, z);
    this.group.add(pillarGroup);
  }

  private createCeilingStructure() {
    // Add some overhead structural beams
    const beamGeometry = new THREE.BoxGeometry(100, 0.5, 0.5);
    const beamMaterial = new THREE.MeshStandardMaterial({
      color: 0x2a2a3a,
      roughness: 0.8,
      metalness: 0.6,
    });

    for (let z = -40; z <= 40; z += 20) {
      const beam = new THREE.Mesh(beamGeometry, beamMaterial);
      beam.position.set(0, 12, z);
      beam.castShadow = true;
      this.group.add(beam);
    }

    // Cross beams
    const crossBeamGeometry = new THREE.BoxGeometry(0.5, 0.5, 100);
    for (let x = -40; x <= 40; x += 20) {
      const beam = new THREE.Mesh(crossBeamGeometry, beamMaterial);
      beam.position.set(x, 12, 0);
      beam.castShadow = true;
      this.group.add(beam);
    }
  }

  private createOverheadLighting() {
    // Industrial overhead lights
    const positions = [
      { x: -20, z: -20 },
      { x: 0, z: -20 },
      { x: 20, z: -20 },
      { x: -20, z: 0 },
      { x: 0, z: 0 },
      { x: 20, z: 0 },
      { x: -20, z: 20 },
      { x: 0, z: 20 },
      { x: 20, z: 20 },
    ];

    positions.forEach((pos) => {
      // Light fixture
      const fixtureGeometry = new THREE.CylinderGeometry(0.8, 0.6, 0.5, 16);
      const fixtureMaterial = new THREE.MeshStandardMaterial({
        color: 0x444444,
        roughness: 0.7,
        metalness: 0.5,
      });
      const fixture = new THREE.Mesh(fixtureGeometry, fixtureMaterial);
      fixture.position.set(pos.x, 11.5, pos.z);
      this.group.add(fixture);

      // Light glow
      const glowGeometry = new THREE.CircleGeometry(0.7, 16);
      const glowMaterial = new THREE.MeshBasicMaterial({
        color: 0xffffaa,
        transparent: true,
        opacity: 0.6,
      });
      const glow = new THREE.Mesh(glowGeometry, glowMaterial);
      glow.position.set(pos.x, 11.2, pos.z);
      glow.rotation.x = -Math.PI / 2;
      this.group.add(glow);

      // Actual light source
      const light = new THREE.SpotLight(0xffffdd, 1, 30, Math.PI / 4, 0.5);
      light.position.set(pos.x, 11, pos.z);
      light.target.position.set(pos.x, 0, pos.z);
      light.castShadow = true;
      light.shadow.mapSize.width = 1024;
      light.shadow.mapSize.height = 1024;
      this.group.add(light);
      this.group.add(light.target);
    });
  }

  private createCableTraySystem() {
    // Overhead cable trays running between regions
    const trayGeometry = new THREE.BoxGeometry(60, 0.2, 1);
    const trayMaterial = new THREE.MeshStandardMaterial({
      color: 0x666677,
      roughness: 0.8,
      metalness: 0.4,
    });

    const trays = [
      { x: 0, y: 10, z: -20 },
      { x: 0, y: 10, z: 0 },
      { x: 0, y: 10, z: 20 },
    ];

    trays.forEach((pos) => {
      const tray = new THREE.Mesh(trayGeometry, trayMaterial);
      tray.position.set(pos.x, pos.y, pos.z);
      tray.castShadow = true;
      this.group.add(tray);

      // Add cable tray mesh/grid
      const meshGeometry = new THREE.PlaneGeometry(60, 1);
      const meshMaterial = new THREE.MeshBasicMaterial({
        color: 0x444455,
        transparent: true,
        opacity: 0.3,
        wireframe: true,
      });
      const mesh = new THREE.Mesh(meshGeometry, meshMaterial);
      mesh.position.set(pos.x, pos.y - 0.2, pos.z);
      mesh.rotation.x = -Math.PI / 2;
      this.group.add(mesh);
    });
  }

  public getGroup(): THREE.Group {
    return this.group;
  }
}
