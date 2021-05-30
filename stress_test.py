#!/bin/env python3

# stress test automation
#
# How to push images from directory
# python3 stress_test.py -action push -workload-id test -token token -frames-path frames
#
# How to pull result images from API
# python3 stress_test.py -action pull -workload-id test -image-type original -token token -frames-path frames
#
# TODO
#
# - Add timing metrics
# -


import argparse
import glob
import os
import requests

WORKLOADS_API_ENDPOINT='http://localhost:8080/workloads'
IMAGES_API_ENDPOINT='http://localhost:8080/images'


# push_images sends images to the API to be filtered.
# images (frames) are taken from an specified directory
def push_images(frames_path, workload_id, token):

    if not os.path.isdir(frames_path):
        print('[{}] frames path doesn\'t exist'.format(frames_path))
        return

    frames = glob.glob('{}/*.png'.format(frames_path))

    data = {'workload_id':workload_id, 'type': 'original'}
    headers= {'Authorization': 'Bearer {}'.format(token)}

    for count in range(0,len(frames)):
        image_path = '{}/{}.png'.format(frames_path,count)

        files = {'data': ('{}.png'.format(count), open(image_path,'rb'), 'image/png')}
        data = {'workload_id': workload_id}

        print('Sending {} frame'.format(image_path))
        r = requests.post(IMAGES_API_ENDPOINT, files=files, headers=headers, data=data)
        print(IMAGES_API_ENDPOINT, data)
        print(r.status_code)


# pull_images pulls results images
def pull_images(frames_path, workload_id, image_type, token):
    if not os.path.isdir(frames_path):
        os.mkdir(frames_path)

    headers= {'Authorization': 'Bearer {}'.format(token)}

    # Get images ids
    get_workload_url = '{}/{}'.format(WORKLOADS_API_ENDPOINT, workload_id)
    r = requests.get(get_workload_url, headers=headers)
    images_ids = r.json()['filtered_images']
    print(images_ids)


    for image in images_ids:
        images_url = '{}/{}'.format(IMAGES_API_ENDPOINT, image)
        image_path = '{}/{}.png'.format(frames_path, image)
        data = {'workload_id':workload_id, 'type':image_type}
        print(images_url, data)
        r = requests.get(images_url, allow_redirects=True, data=data, headers=headers)
        if r.status_code == 200:
            open(image_path, 'wb').write(r.content)



if __name__ == '__main__':

    parser = argparse.ArgumentParser()
    parser.add_argument('-action', default='extract', help='extract or join video frames')
    parser.add_argument('-workload-id', default='test', help='Workload identifier')
    parser.add_argument('-image-type', default='filtered', help='filtered or original')
    parser.add_argument('-token', default='token', help='API Token')
    parser.add_argument('-frames-path', default='frames', help='frames path')

    args = parser.parse_args()
    if args.action == 'push':
        push_images(args.frames_path, args.workload_id, args.token)
    elif args.action == 'pull':
        pull_images(args.frames_path, args.workload_id, args.image_type, args.token)