# vlg
Plex Virtual Library Builder


# Config

vlg.json
```
{
    "hostname": "http://PLEX__HOST",
    "port": 32400,
    "plextoken": "YOUR__PLEX__TOKEN",
    "sections": [
        {
            "_id": "1",
            "name": "Movies",
            "virtuallibpath": "PATH__TO__VIRTUAL__LIBRARY__ROOT__PATH",
            "virtuallibpoolroot": "PATH__TO__POOLED__VIRTUAL__LIBRARY__ROOT",
            "virtuallibpool": [
                "POOLED__DRIVE__1", 
                "POOLED__DRIVE__2"
            ]
        }
    ]
}
```

Where:  
__PLEX__HOST__ = the host name of your Plex Server (http://localhost)  
__YOUR__PLEX__TOKEN__ = your Plex Api Token  

VLG supports two types of Virtual Libraries:  
 * Where the main library and the virtual library are on the same disk or filesystem  
 * Where the main library and the virtual library are part of a Pool like Stablebit Drive Pool  

If your main library and virtual library are on the same file system/disk use this variable:  
__PATH__TO__VIRTUAL__LIBRARY__ROOT__PATH__ would be the path on the disk to where you want your virtual library to reside. (Z:\\Media\\VirtualLib\\)

If your main library and virtual library are part of a pool like Stablebit Drive Pool use these variables:  
__PATH__TO__POOLED__VIRTUAL__LIBRARY__ROOT__ would be the root drive for the pool (Z:)  
__POOLED_DRIVE__1__ would be the first drive in the pool (D:\\PoolPart.2807d48d-46e3-450a-8f23-6f32328a9cd1)  
__POOLED_DRIVE__2__ would be the second drive in the pool (E:\\PoolPart.2807d48d-46e3-450a-8f23-6f32328a9cd1)  
  
  
You can have as many pooled drives as you would like but it is best to have a mapping for each drive in the pool. If not, sometimes the folder will appear empty and Plex wont be able to import the media.

Now, keep in mind when using the Stablebit Drive Pool, you only need to include the roots to the pool and NOT the paths to where the media actually lives. VLG will extract the paths from what Plex provides and generate virtual paths and build the symlinks needed for Plex.

# How to add media to Virtual Libraries

Adding media to virtual libraries is very easy and media can be in as many or as few virtual libraries as you want.

In order to add a title to a virtual library:
* Open plex and find the title you would like to add to a virtual library.
* Click the Edit button (looks like a pencil)
* Click the *Tags* tab.
* Add a virtual library tag to the *Collections* area near the bottom. VLG looks for *VL-* tags (VL-Kids, VL-StandUp, VL-Christmas) VLG supports having titles in multiple virtual libraries. So having VL-Kids VL-Animation on Big Buck Bunny would include it in the Kids and Animation virtual libraries. 
* Click *SAVE CHANGES*
