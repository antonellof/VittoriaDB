"""
Real-time Notification System using WebSockets
Handles stats updates, processing status, and health notifications
"""

import asyncio
import json
import logging
import time
from typing import Dict, List, Any, Optional, Set
from dataclasses import dataclass, asdict
from enum import Enum
from fastapi import WebSocket, WebSocketDisconnect

logger = logging.getLogger(__name__)

class NotificationType(Enum):
    """Types of notifications"""
    STATS_UPDATE = "stats_update"
    PROCESSING_START = "processing_start"
    PROCESSING_PROGRESS = "processing_progress"
    PROCESSING_COMPLETE = "processing_complete"
    PROCESSING_ERROR = "processing_error"
    HEALTH_UPDATE = "health_update"
    COLLECTION_UPDATE = "collection_update"
    SYSTEM_STATUS = "system_status"

@dataclass
class Notification:
    """Notification message structure"""
    type: NotificationType
    data: Dict[str, Any]
    timestamp: float
    id: Optional[str] = None

class WebSocketManager:
    """Manages WebSocket connections and broadcasts notifications"""
    
    def __init__(self):
        self.active_connections: Set[WebSocket] = set()
        self.connection_info: Dict[WebSocket, Dict[str, Any]] = {}
        
    async def connect(self, websocket: WebSocket, client_info: Dict[str, Any] = None):
        """Accept a new WebSocket connection"""
        await websocket.accept()
        self.active_connections.add(websocket)
        self.connection_info[websocket] = client_info or {}
        
        logger.info(f"WebSocket connected: {len(self.active_connections)} total connections")
        
        # Send initial connection confirmation
        await self.send_to_connection(websocket, Notification(
            type=NotificationType.SYSTEM_STATUS,
            data={
                "status": "connected",
                "message": "WebSocket connection established",
                "server_time": time.time()
            },
            timestamp=time.time()
        ))
    
    def disconnect(self, websocket: WebSocket):
        """Remove a WebSocket connection"""
        if websocket in self.active_connections:
            self.active_connections.remove(websocket)
            self.connection_info.pop(websocket, None)
            logger.info(f"WebSocket disconnected: {len(self.active_connections)} total connections")
    
    async def send_to_connection(self, websocket: WebSocket, notification: Notification):
        """Send notification to a specific connection"""
        try:
            message = {
                "type": notification.type.value,
                "data": notification.data,
                "timestamp": notification.timestamp,
                "id": notification.id
            }
            await websocket.send_text(json.dumps(message))
        except Exception as e:
            logger.warning(f"Failed to send message to WebSocket: {e}")
            self.disconnect(websocket)
    
    async def broadcast(self, notification: Notification):
        """Broadcast notification to all connected clients"""
        if not self.active_connections:
            return
        
        logger.debug(f"Broadcasting {notification.type.value} to {len(self.active_connections)} connections")
        
        # Send to all connections concurrently
        tasks = [
            self.send_to_connection(websocket, notification)
            for websocket in self.active_connections.copy()
        ]
        
        if tasks:
            await asyncio.gather(*tasks, return_exceptions=True)

class NotificationService:
    """Service for managing and sending notifications"""
    
    def __init__(self):
        self.websocket_manager = WebSocketManager()
        self.last_stats = None
        self.processing_status: Dict[str, Dict[str, Any]] = {}
        
    async def connect_websocket(self, websocket: WebSocket):
        """Connect a new WebSocket client"""
        await self.websocket_manager.connect(websocket)
        
        # Send current stats if available
        if self.last_stats:
            await self.notify_stats_update(self.last_stats)
    
    def disconnect_websocket(self, websocket: WebSocket):
        """Disconnect a WebSocket client"""
        self.websocket_manager.disconnect(websocket)
    
    async def notify_stats_update(self, stats: Dict[str, Any]):
        """Notify about stats update"""
        self.last_stats = stats
        await self.websocket_manager.broadcast(Notification(
            type=NotificationType.STATS_UPDATE,
            data=stats,
            timestamp=time.time()
        ))
    
    async def notify_processing_start(self, document_id: str, filename: str, file_size: int):
        """Notify about document processing start"""
        self.processing_status[document_id] = {
            "status": "processing",
            "filename": filename,
            "file_size": file_size,
            "start_time": time.time(),
            "progress": 0
        }
        
        await self.websocket_manager.broadcast(Notification(
            type=NotificationType.PROCESSING_START,
            data={
                "document_id": document_id,
                "filename": filename,
                "file_size": file_size,
                "status": "processing"
            },
            timestamp=time.time(),
            id=document_id
        ))
    
    async def notify_processing_progress(self, document_id: str, progress: int, message: str = ""):
        """Notify about document processing progress"""
        if document_id in self.processing_status:
            self.processing_status[document_id]["progress"] = progress
            
        await self.websocket_manager.broadcast(Notification(
            type=NotificationType.PROCESSING_PROGRESS,
            data={
                "document_id": document_id,
                "progress": progress,
                "message": message
            },
            timestamp=time.time(),
            id=document_id
        ))
    
    async def notify_processing_complete(self, document_id: str, chunks_created: int, processing_time: float):
        """Notify about document processing completion"""
        if document_id in self.processing_status:
            self.processing_status[document_id].update({
                "status": "completed",
                "chunks_created": chunks_created,
                "processing_time": processing_time,
                "end_time": time.time()
            })
        
        await self.websocket_manager.broadcast(Notification(
            type=NotificationType.PROCESSING_COMPLETE,
            data={
                "document_id": document_id,
                "status": "completed",
                "chunks_created": chunks_created,
                "processing_time": processing_time
            },
            timestamp=time.time(),
            id=document_id
        ))
    
    async def notify_processing_error(self, document_id: str, error: str):
        """Notify about document processing error"""
        if document_id in self.processing_status:
            self.processing_status[document_id].update({
                "status": "error",
                "error": error,
                "end_time": time.time()
            })
        
        await self.websocket_manager.broadcast(Notification(
            type=NotificationType.PROCESSING_ERROR,
            data={
                "document_id": document_id,
                "status": "error",
                "error": error
            },
            timestamp=time.time(),
            id=document_id
        ))
    
    async def notify_collection_update(self, collection_name: str, stats: Dict[str, Any]):
        """Notify about collection update"""
        await self.websocket_manager.broadcast(Notification(
            type=NotificationType.COLLECTION_UPDATE,
            data={
                "collection_name": collection_name,
                "stats": stats
            },
            timestamp=time.time()
        ))
    
    async def notify_health_update(self, health_status: Dict[str, Any]):
        """Notify about health status change"""
        await self.websocket_manager.broadcast(Notification(
            type=NotificationType.HEALTH_UPDATE,
            data=health_status,
            timestamp=time.time()
        ))
    
    def get_processing_status(self, document_id: str) -> Optional[Dict[str, Any]]:
        """Get processing status for a document"""
        return self.processing_status.get(document_id)
    
    def get_all_processing_status(self) -> Dict[str, Dict[str, Any]]:
        """Get all processing statuses"""
        return self.processing_status.copy()

# Global notification service instance
_notification_service = None

def get_notification_service() -> NotificationService:
    """Get or create global notification service instance"""
    global _notification_service
    if _notification_service is None:
        _notification_service = NotificationService()
    return _notification_service
